package ycfg

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/AmarsDing/lib/container/yarray"
	"github.com/AmarsDing/lib/container/ymap"
	"github.com/AmarsDing/lib/encoding/yjson"
	"github.com/AmarsDing/lib/errors/yerror"
	"github.com/AmarsDing/lib/internal/intlog"
	"github.com/AmarsDing/lib/os/ycmd"
	"github.com/AmarsDing/lib/os/yfile"
	"github.com/AmarsDing/lib/os/yfsnotify"
	"github.com/AmarsDing/lib/os/ylog"
	"github.com/AmarsDing/lib/os/yres"
	"github.com/AmarsDing/lib/os/yspath"
	"github.com/AmarsDing/lib/text/ystr"
	"github.com/AmarsDing/lib/util/ymode"
)

// New returns a new configuration management object.
// The parameter `file` specifies the default configuration file name for reading.
func New(file ...string) *Config {
	name := DefaultConfiyfile
	if len(file) > 0 {
		name = file[0]
	} else {
		// Custom default configuration file name from command line or environment.
		if customFile := ycmd.GetOptWithEnv(fmt.Sprintf("%s.file", cmdEnvKey)).String(); customFile != "" {
			name = customFile
		}
	}
	c := &Config{
		defaultName: name,
		searchPaths: yarray.NewStrArray(true),
		jsonMap:     ymap.NewStrAnyMap(true),
	}
	// Customized dir path from env/cmd.
	if customPath := ycmd.GetOptWithEnv(fmt.Sprintf("%s.path", cmdEnvKey)).String(); customPath != "" {
		if yfile.Exists(customPath) {
			_ = c.SetPath(customPath)
		} else {
			if errorPrint() {
				ylog.Errorf("[gcfg] Configuration directory path does not exist: %s", customPath)
			}
		}
	} else {
		// Dir path of working dir.
		if err := c.AddPath(yfile.Pwd()); err != nil {
			intlog.Error(err)
		}

		// Dir path of main package.
		if mainPath := yfile.MainPkgPath(); mainPath != "" && yfile.Exists(mainPath) {
			if err := c.AddPath(mainPath); err != nil {
				intlog.Error(err)
			}
		}

		// Dir path of binary.
		if selfPath := yfile.SelfDir(); selfPath != "" && yfile.Exists(selfPath) {
			if err := c.AddPath(selfPath); err != nil {
				intlog.Error(err)
			}
		}
	}
	return c
}

// Instance returns an instance of Config with default settings.
// The parameter `name` is the name for the instance. But very note that, if the file "name.toml"
// exists in the configuration directory, it then sets it as the default configuration file. The
// toml file type is the default configuration file type.
func Instance(name ...string) *Config {
	key := DefaultName
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}
	return instances.GetOrSetFuncLock(key, func() interface{} {
		c := New()
		// If it's not using default configuration or its configuration file is not available,
		// it searches the possible configuration file according to the name and all supported
		// file types.
		if key != DefaultName || !c.Available() {
			for _, fileType := range supportedFileTypes {
				if file := fmt.Sprintf(`%s.%s`, key, fileType); c.Available(file) {
					c.SetFileName(file)
					break
				}
			}
		}
		return c
	}).(*Config)
}

// SetPath sets the configuration directory path for file search.
// The parameter `path` can be absolute or relative path,
// but absolute path is strongly recommended.
func (c *Config) SetPath(path string) error {
	var (
		isDir    = false
		realPath = ""
	)
	if file := yres.Get(path); file != nil {
		realPath = path
		isDir = file.FileInfo().IsDir()
	} else {
		// Absolute path.
		realPath = yfile.RealPath(path)
		if realPath == "" {
			// Relative path.
			c.searchPaths.RLockFunc(func(array []string) {
				for _, v := range array {
					if path, _ := yspath.Search(v, path); path != "" {
						realPath = path
						break
					}
				}
			})
		}
		if realPath != "" {
			isDir = yfile.IsDir(realPath)
		}
	}
	// Path not exist.
	if realPath == "" {
		buffer := bytes.NewBuffer(nil)
		if c.searchPaths.Len() > 0 {
			buffer.WriteString(fmt.Sprintf("[gcfg] SetPath failed: cannot find directory \"%s\" in following paths:", path))
			c.searchPaths.RLockFunc(func(array []string) {
				for k, v := range array {
					buffer.WriteString(fmt.Sprintf("\n%d. %s", k+1, v))
				}
			})
		} else {
			buffer.WriteString(fmt.Sprintf(`[gcfg] SetPath failed: path "%s" does not exist`, path))
		}
		err := errors.New(buffer.String())
		if errorPrint() {
			ylog.Error(err)
		}
		return err
	}
	// Should be a directory.
	if !isDir {
		err := fmt.Errorf(`[gcfg] SetPath failed: path "%s" should be directory type`, path)
		if errorPrint() {
			ylog.Error(err)
		}
		return err
	}
	// Repeated path check.
	if c.searchPaths.Search(realPath) != -1 {
		return nil
	}
	c.jsonMap.Clear()
	c.searchPaths.Clear()
	c.searchPaths.Append(realPath)
	intlog.Print("SetPath:", realPath)
	return nil
}

// SetViolenceCheck sets whether to perform hierarchical conflict checking.
// This feature needs to be enabled when there is a level symbol in the key name.
// It is off in default.
//
// Note that, turning on this feature is quite expensive, and it is not recommended
// to allow separators in the key names. It is best to avoid this on the application side.
func (c *Config) SetViolenceCheck(check bool) {
	c.violenceCheck = check
	c.Clear()
}

// AddPath adds a absolute or relative path to the search paths.
func (c *Config) AddPath(path string) error {
	var (
		isDir    = false
		realPath = ""
	)
	// It firstly checks the resource manager,
	// and then checks the filesystem for the path.
	if file := yres.Get(path); file != nil {
		realPath = path
		isDir = file.FileInfo().IsDir()
	} else {
		// Absolute path.
		realPath = yfile.RealPath(path)
		if realPath == "" {
			// Relative path.
			c.searchPaths.RLockFunc(func(array []string) {
				for _, v := range array {
					if path, _ := yspath.Search(v, path); path != "" {
						realPath = path
						break
					}
				}
			})
		}
		if realPath != "" {
			isDir = yfile.IsDir(realPath)
		}
	}
	if realPath == "" {
		buffer := bytes.NewBuffer(nil)
		if c.searchPaths.Len() > 0 {
			buffer.WriteString(fmt.Sprintf("[gcfg] AddPath failed: cannot find directory \"%s\" in following paths:", path))
			c.searchPaths.RLockFunc(func(array []string) {
				for k, v := range array {
					buffer.WriteString(fmt.Sprintf("\n%d. %s", k+1, v))
				}
			})
		} else {
			buffer.WriteString(fmt.Sprintf(`[gcfg] AddPath failed: path "%s" does not exist`, path))
		}
		err := yerror.New(buffer.String())
		if errorPrint() {
			ylog.Error(err)
		}
		return err
	}
	if !isDir {
		err := yerror.Newf(`[gcfg] AddPath failed: path "%s" should be directory type`, path)
		if errorPrint() {
			ylog.Error(err)
		}
		return err
	}
	// Repeated path check.
	if c.searchPaths.Search(realPath) != -1 {
		return nil
	}
	c.searchPaths.Append(realPath)
	intlog.Print("AddPath:", realPath)
	return nil
}

// SetFileName sets the default configuration file name.
func (c *Config) SetFileName(name string) *Config {
	c.defaultName = name
	return c
}

// GetFileName returns the default configuration file name.
func (c *Config) GetFileName() string {
	return c.defaultName
}

// Available checks and returns whether configuration of given `file` is available.
func (c *Config) Available(file ...string) bool {
	var name string
	if len(file) > 0 && file[0] != "" {
		name = file[0]
	} else {
		name = c.defaultName
	}
	if path, _ := c.GetFilePath(name); path != "" {
		return true
	}
	if GetContent(name) != "" {
		return true
	}
	return false
}

// GetFilePath returns the absolute configuration file path for the given filename by `file`.
// If `file` is not passed, it returns the configuration file path of the default name.
// It returns an empty `path` string and an error if the given `file` does not exist.
func (c *Config) GetFilePath(file ...string) (path string, err error) {
	name := c.defaultName
	if len(file) > 0 {
		name = file[0]
	}
	// Searching resource manager.
	if !yres.IsEmpty() {
		for _, v := range resourceTryFiles {
			if file := yres.Get(v + name); file != nil {
				path = file.Name()
				return
			}
		}
		c.searchPaths.RLockFunc(func(array []string) {
			for _, prefix := range array {
				for _, v := range resourceTryFiles {
					if file := yres.Get(prefix + v + name); file != nil {
						path = file.Name()
						return
					}
				}
			}
		})
	}
	c.autoCheckAndAddMainPkgPathToSearchPaths()
	// Searching the file system.
	c.searchPaths.RLockFunc(func(array []string) {
		for _, prefix := range array {
			prefix = ystr.TrimRight(prefix, `\/`)
			if path, _ = yspath.Search(prefix, name); path != "" {
				return
			}
			if path, _ = yspath.Search(prefix+yfile.Separator+"config", name); path != "" {
				return
			}
		}
	})
	// If it cannot find the path of `file`, it formats and returns a detailed error.
	if path == "" {
		var (
			buffer = bytes.NewBuffer(nil)
		)
		if c.searchPaths.Len() > 0 {
			buffer.WriteString(fmt.Sprintf(`[gcfg] cannot find config file "%s" in resource manager or the following paths:`, name))
			c.searchPaths.RLockFunc(func(array []string) {
				index := 1
				for _, v := range array {
					v = ystr.TrimRight(v, `\/`)
					buffer.WriteString(fmt.Sprintf("\n%d. %s", index, v))
					index++
					buffer.WriteString(fmt.Sprintf("\n%d. %s", index, v+yfile.Separator+"config"))
					index++
				}
			})
		} else {
			buffer.WriteString(fmt.Sprintf("[gcfg] cannot find config file \"%s\" with no path configured", name))
		}
		err = yerror.New(buffer.String())
	}
	return
}

// autoCheckAndAddMainPkgPathToSearchPaths automatically checks and adds directory path of package main
// to the searching path list if it's currently in development environment.
func (c *Config) autoCheckAndAddMainPkgPathToSearchPaths() {
	if ymode.IsDevelop() {
		mainPkgPath := yfile.MainPkgPath()
		if mainPkgPath != "" {
			if !c.searchPaths.Contains(mainPkgPath) {
				c.searchPaths.Append(mainPkgPath)
			}
		}
	}
}

// getJson returns a *yjson.Json object for the specified `file` content.
// It would print error if file reading fails. It return nil if any error occurs.
func (c *Config) getJson(file ...string) *yjson.Json {
	var name string
	if len(file) > 0 && file[0] != "" {
		name = file[0]
	} else {
		name = c.defaultName
	}
	r := c.jsonMap.GetOrSetFuncLock(name, func() interface{} {
		var (
			err      error
			content  string
			filePath string
		)
		// The configured content can be any kind of data type different from its file type.
		isFromConfigContent := true
		if content = GetContent(name); content == "" {
			isFromConfigContent = false
			filePath, err = c.GetFilePath(name)
			if err != nil && errorPrint() {
				ylog.Error(err)
			}
			if filePath == "" {
				return nil
			}
			if file := yres.Get(filePath); file != nil {
				content = string(file.Content())
			} else {
				content = yfile.GetContents(filePath)
			}
		}
		// Note that the underlying configuration json object operations are concurrent safe.
		var (
			j *yjson.Json
		)
		dataType := yfile.ExtName(name)
		if yjson.IsValidDataType(dataType) && !isFromConfigContent {
			j, err = yjson.LoadContentType(dataType, content, true)
		} else {
			j, err = yjson.LoadContent(content, true)
		}
		if err == nil {
			j.SetViolenceCheck(c.violenceCheck)
			// Add monitor for this configuration file,
			// any changes of this file will refresh its cache in Config object.
			if filePath != "" && !yres.Contains(filePath) {
				_, err = yfsnotify.Add(filePath, func(event *yfsnotify.Event) {
					c.jsonMap.Remove(name)
				})
				if err != nil && errorPrint() {
					ylog.Error(err)
				}
			}
			return j
		}
		if errorPrint() {
			if filePath != "" {
				ylog.Criticalf(`[gcfg] load config file "%s" failed: %s`, filePath, err.Error())
			} else {
				ylog.Criticalf(`[gcfg] load configuration failed: %s`, err.Error())
			}
		}
		return nil
	})
	if r != nil {
		return r.(*yjson.Json)
	}
	return nil
}
