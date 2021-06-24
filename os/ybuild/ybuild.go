package ybuild

import (
	"runtime"

	"github.com/AmarsDing/lib"
	"github.com/AmarsDing/lib/container/yvar"
	"github.com/AmarsDing/lib/encoding/ybase64"
	"github.com/AmarsDing/lib/internal/intlog"
	"github.com/AmarsDing/lib/internal/json"
	"github.com/AmarsDing/lib/util/yconv"
)

var (
	builtInVarStr = ""                       // Raw variable base64 string.
	builtInVarMap = map[string]interface{}{} // Binary custom variable map decoded.
)

func init() {
	if builtInVarStr != "" {
		err := json.UnmarshalUseNumber(ybase64.MustDecodeString(builtInVarStr), &builtInVarMap)
		if err != nil {
			intlog.Error(err)
		}
		builtInVarMap["libVersion"] = lib.VERSION
		builtInVarMap["goVersion"] = runtime.Version()
		intlog.Printf("build variables: %+v", builtInVarMap)
	} else {
		intlog.Print("no build variables")
	}
}

// Info returns the basic built information of the binary as map.
// Note that it should be used with gf-cli tool "gf build",
// which injects necessary information into the binary.
func Info() map[string]string {
	return map[string]string{
		"lib":  GetString("libVersion"),
		"go":   GetString("goVersion"),
		"git":  GetString("builtGit"),
		"time": GetString("builtTime"),
	}
}

// Get retrieves and returns the build-in binary variable with given name.
func Get(name string, def ...interface{}) interface{} {
	if v, ok := builtInVarMap[name]; ok {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return nil
}

// GetVar retrieves and returns the build-in binary variable of given name as yvar.Var.
func GetVar(name string, def ...interface{}) *yvar.Var {
	return yvar.New(Get(name, def...))
}

// GetString retrieves and returns the build-in binary variable of given name as string.
func GetString(name string, def ...interface{}) string {
	return yconv.String(Get(name, def...))
}

// Map returns the custom build-in variable map.
func Map() map[string]interface{} {
	return builtInVarMap
}
