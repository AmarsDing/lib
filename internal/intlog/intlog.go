package intlog

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/AmarsDing/lib/debug/ydebug"
	"github.com/AmarsDing/lib/internal/utils"
)

const (
	stackFilterKey = "/internal/intlog"
)

var (
	// islibDebug marks whether printing GoFrame debug information.
	islibDebug = false
)

func init() {
	islibDebug = utils.IsDebugEnabled()
}

// SetEnabled enables/disables the internal logging manually.
// Note that this function is not concurrent safe, be aware of the DATA RACE.
func SetEnabled(enabled bool) {
	// If they're the same, it does not write the `islibDebug` but only reading operation.
	if islibDebug != enabled {
		islibDebug = enabled
	}
}

// Print prints `v` with newline using fmt.Println.
// The parameter `v` can be multiple variables.
func Print(v ...interface{}) {
	if !islibDebug {
		return
	}
	fmt.Println(append([]interface{}{now(), "[INTE]", file()}, v...)...)
}

// Printf prints `v` with format `format` using fmt.Printf.
// The parameter `v` can be multiple variables.
func Printf(format string, v ...interface{}) {
	if !islibDebug {
		return
	}
	fmt.Printf(now()+" [INTE] "+file()+" "+format+"\n", v...)
}

// Error prints `v` with newline using fmt.Println.
// The parameter `v` can be multiple variables.
func Error(v ...interface{}) {
	if !islibDebug {
		return
	}
	array := append([]interface{}{now(), "[INTE]", file()}, v...)
	array = append(array, "\n"+ydebug.StackWithFilter(stackFilterKey))
	fmt.Println(array...)
}

// Errorf prints `v` with format `format` using fmt.Printf.
func Errorf(format string, v ...interface{}) {
	if !islibDebug {
		return
	}
	fmt.Printf(
		now()+" [INTE] "+file()+" "+format+"\n%s\n",
		append(v, ydebug.StackWithFilter(stackFilterKey))...,
	)
}

// now returns current time string.
func now() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

// file returns caller file name along with its line number.
func file() string {
	_, p, l := ydebug.CallerWithFilter(stackFilterKey)
	return fmt.Sprintf(`%s:%d`, filepath.Base(p), l)
}
