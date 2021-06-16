package utils

import (
	"github.com/AmarsDing/lib/internal/command"
)

const (
	debugKey                 = "lib.debug"                  // Debug key for checking if in debug mode.
	StackFilterKeyForGoFrame = "/github.com/AmarsDing/lib/" // Stack filtering key for all GoFrame module paths.
)

var (
	// isDebugEnabled marks whether GoFrame debug mode is enabled.
	isDebugEnabled = false
)

func init() {
	// Debugging configured.
	value := command.GetOptWithEnv(debugKey)
	if value == "" || value == "0" || value == "false" {
		isDebugEnabled = false
	} else {
		isDebugEnabled = true
	}
}

// IsDebugEnabled checks and returns whether debug mode is enabled.
// The debug mode is enabled when command argument "gf.debug" or environment "GF_DEBUG" is passed.
func IsDebugEnabled() bool {
	return isDebugEnabled
}
