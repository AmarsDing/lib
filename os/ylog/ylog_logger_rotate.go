package ylog

import (
	"fmt"
	"time"

	"github.com/AmarsDing/lib/container/yarray"
	"github.com/AmarsDing/lib/encoding/ycompress"
	"github.com/AmarsDing/lib/internal/intlog"
	"github.com/AmarsDing/lib/os/yfile"
	"github.com/AmarsDing/lib/os/ymlock"
	"github.com/AmarsDing/lib/os/ytime"
	"github.com/AmarsDing/lib/os/ytimer"
	"github.com/AmarsDing/lib/text/yregex"
)

// rotateFileBySize rotates the current logging file according to the
// configured rotation size.
func (l *Logger) rotateFileBySize(now time.Time) {
	if l.config.RotateSize <= 0 {
		return
	}
	if err := l.doRotateFile(l.getFilePath(now)); err != nil {
		// panic(err)
		intlog.Error(err)
	}
}

// doRotateFile rotates the given logging file.
func (l *Logger) doRotateFile(filePath string) error {
	memoryLockKey := "glog.doRotateFile:" + filePath
	if !ymlock.TryLock(memoryLockKey) {
		return nil
	}
	defer ymlock.Unlock(memoryLockKey)

	// No backups, it then just removes the current logging file.
	if l.config.RotateBackupLimit == 0 {
		if err := yfile.Remove(filePath); err != nil {
			return err
		}
		intlog.Printf(`%d size exceeds, no backups set, remove original logging file: %s`, l.config.RotateSize, filePath)
		return nil
	}
	// Else it creates new backup files.
	var (
		dirPath     = yfile.Dir(filePath)
		fileName    = yfile.Name(filePath)
		fileExtName = yfile.ExtName(filePath)
		newFilePath = ""
	)
	// Rename the logging file by adding extra datetime information to microseconds, like:
	// access.log          -> access.20200326101301899002.log
	// access.20200326.log -> access.20200326.20200326101301899002.log
	for {
		var (
			now   = ytime.Now()
			micro = now.Microsecond() % 1000
		)
		if micro == 0 {
			micro = 101
		} else {
			for micro < 100 {
				micro *= 10
			}
		}
		newFilePath = yfile.Join(
			dirPath,
			fmt.Sprintf(
				`%s.%s%d.%s`,
				fileName, now.Format("YmdHisu"), micro, fileExtName,
			),
		)
		if !yfile.Exists(newFilePath) {
			break
		} else {
			intlog.Printf(`rotation file exists, continue: %s`, newFilePath)
		}
	}
	if err := yfile.Rename(filePath, newFilePath); err != nil {
		return err
	}
	return nil
}

// rotateChecksTimely timely checks the backups expiration and the compression.
func (l *Logger) rotateChecksTimely() {
	defer ytimer.AddOnce(l.config.RotateCheckInterval, l.rotateChecksTimely)
	// Checks whether file rotation not enabled.
	if l.config.RotateSize <= 0 && l.config.RotateExpire == 0 {
		intlog.Printf(
			"logging rotation ignore checks: RotateSize: %d, RotateExpire: %s",
			l.config.RotateSize, l.config.RotateExpire.String(),
		)
		return
	}

	// It here uses memory lock to guarantee the concurrent safety.
	memoryLockKey := "glog.rotateChecksTimely:" + l.config.Path
	if !ymlock.TryLock(memoryLockKey) {
		return
	}
	defer ymlock.Unlock(memoryLockKey)

	var (
		now      = time.Now()
		pattern  = "*.log, *.gz"
		files, _ = yfile.ScanDirFile(l.config.Path, pattern, true)
	)
	intlog.Printf("logging rotation start checks: %+v", files)
	// =============================================================
	// Rotation of expired file checks.
	// =============================================================
	if l.config.RotateExpire > 0 {
		var (
			mtime         time.Time
			subDuration   time.Duration
			expireRotated bool
		)
		for _, file := range files {
			if yfile.ExtName(file) == "gz" {
				continue
			}
			mtime = yfile.MTime(file)
			subDuration = now.Sub(mtime)
			if subDuration > l.config.RotateExpire {
				expireRotated = true
				intlog.Printf(
					`%v - %v = %v > %v, rotation expire logging file: %s`,
					now, mtime, subDuration, l.config.RotateExpire, file,
				)
				if err := l.doRotateFile(file); err != nil {
					intlog.Error(err)
				}
			}
		}
		if expireRotated {
			// Update the files array.
			files, _ = yfile.ScanDirFile(l.config.Path, pattern, true)
		}
	}

	// =============================================================
	// Rotated file compression.
	// =============================================================
	needCompressFileArray := yarray.NewStrArray()
	if l.config.RotateBackupCompress > 0 {
		for _, file := range files {
			// Eg: access.20200326101301899002.log.gz
			if yfile.ExtName(file) == "gz" {
				continue
			}
			// Eg:
			// access.20200326101301899002.log
			if yregex.IsMatchString(`.+\.\d{20}\.log`, yfile.Basename(file)) {
				needCompressFileArray.Append(file)
			}
		}
		if needCompressFileArray.Len() > 0 {
			needCompressFileArray.Iterator(func(_ int, path string) bool {
				err := ycompress.GzipFile(path, path+".gz")
				if err == nil {
					intlog.Printf(`compressed done, remove original logging file: %s`, path)
					if err = yfile.Remove(path); err != nil {
						intlog.Print(err)
					}
				} else {
					intlog.Print(err)
				}
				return true
			})
			// Update the files array.
			files, _ = yfile.ScanDirFile(l.config.Path, pattern, true)
		}
	}

	// =============================================================
	// Backups count limitation and expiration checks.
	// =============================================================
	var (
		backupFilesMap          = make(map[string]*yarray.SortedArray)
		originalLogginyfilePath = ""
	)
	if l.config.RotateBackupLimit > 0 || l.config.RotateBackupExpire > 0 {
		for _, file := range files {
			originalLogginyfilePath, _ = yregex.ReplaceString(`\.\d{20}`, "", file)
			if backupFilesMap[originalLogginyfilePath] == nil {
				backupFilesMap[originalLogginyfilePath] = yarray.NewSortedArray(func(a, b interface{}) int {
					// Sorted by rotated/backup file mtime.
					// The old rotated/backup file is put in the head of array.
					file1 := a.(string)
					file2 := b.(string)
					result := yfile.MTimestampMilli(file1) - yfile.MTimestampMilli(file2)
					if result <= 0 {
						return -1
					}
					return 1
				})
			}
			// Check if this file a rotated/backup file.
			if yregex.IsMatchString(`.+\.\d{20}\.log`, yfile.Basename(file)) {
				backupFilesMap[originalLogginyfilePath].Add(file)
			}
		}
		intlog.Printf(`calculated backup files map: %+v`, backupFilesMap)
		for _, array := range backupFilesMap {
			diff := array.Len() - l.config.RotateBackupLimit
			for i := 0; i < diff; i++ {
				path, _ := array.PopLeft()
				intlog.Printf(`remove exceeded backup limit file: %s`, path)
				if err := yfile.Remove(path.(string)); err != nil {
					intlog.Print(err)
				}
			}
		}
		// Backup expiration checks.
		if l.config.RotateBackupExpire > 0 {
			var (
				mtime       time.Time
				subDuration time.Duration
			)
			for _, array := range backupFilesMap {
				array.Iterator(func(_ int, v interface{}) bool {
					path := v.(string)
					mtime = yfile.MTime(path)
					subDuration = now.Sub(mtime)
					if subDuration > l.config.RotateBackupExpire {
						intlog.Printf(
							`%v - %v = %v > %v, remove expired backup file: %s`,
							now, mtime, subDuration, l.config.RotateBackupExpire, path,
						)
						if err := yfile.Remove(path); err != nil {
							intlog.Print(err)
						}
						return true
					} else {
						return false
					}
				})
			}
		}
	}
}
