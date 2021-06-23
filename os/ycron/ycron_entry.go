package ycron

import (
	"reflect"
	"runtime"
	"time"

	"github.com/AmarsDing/lib/container/ytype"
	"github.com/AmarsDing/lib/os/ylog"
	"github.com/AmarsDing/lib/os/ytimer"
	"github.com/AmarsDing/lib/util/yconv"
)

// Entry is timing task entry.
type Entry struct {
	cron     *Cron         // Cron object belonged to.
	entry    *ytimer.Entry // Associated ytimer.Entry.
	schedule *cronSchedule // Timed schedule object.
	jobName  string        // Callback function name(address info).
	times    *ytype.Int    // Running times limit.
	Name     string        // Entry name.
	Job      func()        `json:"-"` // Callback function.
	Time     time.Time     // Registered time.
}

// addEntry creates and returns a new Entry object.
// Param <job> is the callback function for timed task execution.
// Param <singleton> specifies whether timed task executing in singleton mode.
// Param <name> names this entry for manual control.
func (c *Cron) addEntry(pattern string, job func(), singleton bool, name ...string) (*Entry, error) {
	schedule, err := newSchedule(pattern)
	if err != nil {
		return nil, err
	}
	// No limit for <times>, for ytimer checking scheduling every second.
	entry := &Entry{
		cron:     c,
		schedule: schedule,
		jobName:  runtime.FuncForPC(reflect.ValueOf(job).Pointer()).Name(),
		times:    ytype.NewInt(defaultTimes),
		Job:      job,
		Time:     time.Now(),
	}
	if len(name) > 0 {
		entry.Name = name[0]
	} else {
		entry.Name = "gcron-" + yconv.String(c.idGen.Add(1))
	}
	// When you add a scheduled task, you cannot allow it to run.
	// It cannot start running when added to ytimer.
	// It should start running after the entry is added to the entries map,
	// to avoid the task from running during adding where the entries
	// does not have the entry information, which might cause panic.
	entry.entry = ytimer.AddEntry(time.Second, entry.check, singleton, -1, ytimer.StatusStopped)
	c.entries.Set(entry.Name, entry)
	entry.entry.Start()
	return entry, nil
}

// IsSingleton return whether this entry is a singleton timed task.
func (entry *Entry) IsSingleton() bool {
	return entry.entry.IsSingleton()
}

// SetSingleton sets the entry running in singleton mode.
func (entry *Entry) SetSingleton(enabled bool) {
	entry.entry.SetSingleton(true)
}

// SetTimes sets the times which the entry can run.
func (entry *Entry) SetTimes(times int) {
	entry.times.Set(times)
}

// Status returns the status of entry.
func (entry *Entry) Status() int {
	return entry.entry.Status()
}

// SetStatus sets the status of the entry.
func (entry *Entry) SetStatus(status int) int {
	return entry.entry.SetStatus(status)
}

// Start starts running the entry.
func (entry *Entry) Start() {
	entry.entry.Start()
}

// Stop stops running the entry.
func (entry *Entry) Stop() {
	entry.entry.Stop()
}

// Close stops and removes the entry from cron.
func (entry *Entry) Close() {
	entry.cron.entries.Remove(entry.Name)
	entry.entry.Close()
}

// Timing task check execution.
// The running times limits feature is implemented by gcron.Entry and cannot be implemented by ytimer.Entry.
// gcron.Entry relies on ytimer to implement a scheduled task check for gcron.Entry per second.
func (entry *Entry) check() {
	if entry.schedule.meet(time.Now()) {
		path := entry.cron.GetLogPath()
		level := entry.cron.GetLogLevel()
		switch entry.cron.status.Val() {
		case StatusStopped:
			return

		case StatusClosed:
			ylog.Path(path).Level(level).Debugf("[gcron] %s(%s) %s removed", entry.Name, entry.schedule.pattern, entry.jobName)
			entry.Close()

		case StatusReady:
			fallthrough
		case StatusRunning:
			// Running times check.
			times := entry.times.Add(-1)
			if times <= 0 {
				if entry.entry.SetStatus(StatusClosed) == StatusClosed || times < 0 {
					return
				}
			}
			if times < 2000000000 && times > 1000000000 {
				entry.times.Set(defaultTimes)
			}
			ylog.Path(path).Level(level).Debugf("[gcron] %s(%s) %s start", entry.Name, entry.schedule.pattern, entry.jobName)
			defer func() {
				if err := recover(); err != nil {
					ylog.Path(path).Level(level).Errorf("[gcron] %s(%s) %s end with error: %v", entry.Name, entry.schedule.pattern, entry.jobName, err)
				} else {
					ylog.Path(path).Level(level).Debugf("[gcron] %s(%s) %s end", entry.Name, entry.schedule.pattern, entry.jobName)
				}
				if entry.entry.Status() == StatusClosed {
					entry.Close()
				}
			}()
			entry.Job()

		}
	}
}
