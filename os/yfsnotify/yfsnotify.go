// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/AmarsDing/lib.

// Package gfsnotify provides a platform-independent interface for file system notifications.
package yfsnotify

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/AmarsDing/lib/container/ylist"
	"github.com/AmarsDing/lib/container/yset"
	"github.com/AmarsDing/lib/internal/intlog"

	"github.com/AmarsDing/lib/container/ymap"
	"github.com/AmarsDing/lib/container/yqueue"
	"github.com/AmarsDing/lib/container/ytype"
	"github.com/AmarsDing/lib/os/ycache"
	"github.com/fsnotify/fsnotify"
)

// Watcher is the monitor for file changes.
type Watcher struct {
	watcher   *fsnotify.Watcher // Underlying fsnotify object.
	events    *yqueue.Queue     // Used for internal event management.
	cache     *ycache.Cache     // Used for repeated event filter.
	nameSet   *yset.StrSet      // Used for AddOnce feature.
	callbacks *ymap.StrAnyMap   // Path(file/folder) to callbacks mapping.
	closeChan chan struct{}     // Used for watcher closing notification.
}

// Callback is the callback function for Watcher.
type Callback struct {
	Id        int                // Unique id for callback object.
	Func      func(event *Event) // Callback function.
	Path      string             // Bound file path (absolute).
	name      string             // Registered name for AddOnce.
	elem      *ylist.Element     // Element in the callbacks of watcher.
	recursive bool               // Is bound to path recursively or not.
}

// Event is the event produced by underlying fsnotify.
type Event struct {
	event   fsnotify.Event // Underlying event.
	Path    string         // Absolute file path.
	Op      Op             // File operation.
	Watcher *Watcher       // Parent watcher.
}

// Op is the bits union for file operations.
type Op uint32

const (
	CREATE Op = 1 << iota
	WRITE
	REMOVE
	RENAME
	CHMOD
)

const (
	repeatEventFilterDuration = time.Millisecond // Duration for repeated event filter.
	callbackExitEventPanicStr = "exit"           // Custom exit event for internal usage.
)

var (
	mu                  sync.Mutex                // Mutex for concurrent safety of defaultWatcher.
	defaultWatcher      *Watcher                  // Default watcher.
	callbackIdMap       = ymap.NewIntAnyMap(true) // Id to callback mapping.
	callbackIdGenerator = ytype.NewInt()          // Atomic id generator for callback.
)

// New creates and returns a new watcher.
// Note that the watcher number is limited by the file handle setting of the system.
// Eg: fs.inotify.max_user_instances system variable in linux systems.
func New() (*Watcher, error) {
	w := &Watcher{
		cache:     ycache.New(),
		events:    yqueue.New(),
		nameSet:   yset.NewStrSet(true),
		closeChan: make(chan struct{}),
		callbacks: ymap.NewStrAnyMap(true),
	}
	if watcher, err := fsnotify.NewWatcher(); err == nil {
		w.watcher = watcher
	} else {
		intlog.Printf("New watcher failed: %v", err)
		return nil, err
	}
	w.watchLoop()
	w.eventLoop()
	return w, nil
}

// Add monitors <path> using default watcher with callback function <callbackFunc>.
// The optional parameter <recursive> specifies whether monitoring the <path> recursively, which is true in default.
func Add(path string, callbackFunc func(event *Event), recursive ...bool) (callback *Callback, err error) {
	w, err := getDefaultWatcher()
	if err != nil {
		return nil, err
	}
	return w.Add(path, callbackFunc, recursive...)
}

// AddOnce monitors <path> using default watcher with callback function <callbackFunc> only once using unique name <name>.
// If AddOnce is called multiple times with the same <name> parameter, <path> is only added to monitor once. It returns error
// if it's called twice with the same <name>.
//
// The optional parameter <recursive> specifies whether monitoring the <path> recursively, which is true in default.
func AddOnce(name, path string, callbackFunc func(event *Event), recursive ...bool) (callback *Callback, err error) {
	w, err := getDefaultWatcher()
	if err != nil {
		return nil, err
	}
	return w.AddOnce(name, path, callbackFunc, recursive...)
}

// Remove removes all monitoring callbacks of given <path> from watcher recursively.
func Remove(path string) error {
	w, err := getDefaultWatcher()
	if err != nil {
		return err
	}
	return w.Remove(path)
}

// RemoveCallback removes specified callback with given id from watcher.
func RemoveCallback(callbackId int) error {
	w, err := getDefaultWatcher()
	if err != nil {
		return err
	}
	callback := (*Callback)(nil)
	if r := callbackIdMap.Get(callbackId); r != nil {
		callback = r.(*Callback)
	}
	if callback == nil {
		return errors.New(fmt.Sprintf(`callback for id %d not found`, callbackId))
	}
	w.RemoveCallback(callbackId)
	return nil
}

// Exit is only used in the callback function, which can be used to remove current callback
// of itself from the watcher.
func Exit() {
	panic(callbackExitEventPanicStr)
}

// getDefaultWatcher creates and returns the default watcher.
// This is used for lazy initialization purpose.
func getDefaultWatcher() (*Watcher, error) {
	mu.Lock()
	defer mu.Unlock()
	if defaultWatcher != nil {
		return defaultWatcher, nil
	}
	var err error
	defaultWatcher, err = New()
	return defaultWatcher, err
}
