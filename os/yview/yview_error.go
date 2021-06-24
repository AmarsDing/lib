package yview

import (
	"github.com/AmarsDing/lib/os/ycmd"
)

const (
	// gERROR_PRINT_KEY is used to specify the key controlling error printing to stdout.
	// This error is designed not to be returned by functions.
	yERROR_PRINT_KEY = "lib.yview.errorprint"
)

// errorPrint checks whether printing error to stdout.
func errorPrint() bool {
	return ycmd.GetOptWithEnv(yERROR_PRINT_KEY, true).Bool()
}
