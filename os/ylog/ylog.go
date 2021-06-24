package ylog

import (
	"github.com/AmarsDing/lib/os/ycmd"
	"github.com/AmarsDing/lib/os/yrpool"
)

var (
	// Default logger object, for package method usage.
	logger = New()

	// Goroutine pool for async logging output.
	// It uses only one asynchronize worker to ensure log sequence.
	asyncPool = yrpool.New(1)

	// defaultDebug enables debug level or not in default,
	// which can be configured using command option or system environment.
	defaultDebug = true
)

func init() {
	defaultDebug = ycmd.GetOptWithEnv("lib.ylog.debug", true).Bool()
	SetDebug(defaultDebug)
}

// DefaultLogger returns the default logger.
func DefaultLogger() *Logger {
	return logger
}

// SetDefaultLogger sets the default logger for package glog.
// Note that there might be concurrent safety issue if calls this function
// in different goroutines.
func SetDefaultLogger(l *Logger) {
	logger = l
}
