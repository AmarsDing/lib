package ycfg

import (
	"github.com/AmarsDing/lib/container/yarray"
	"github.com/AmarsDing/lib/container/ymap"
	"github.com/AmarsDing/lib/internal/intlog"
	"github.com/AmarsDing/lib/os/ycmd"
)

// Config is the configuration manager.
type Config struct {
	defaultName   string           // Default configuration file name.
	searchPaths   *yarray.StrArray // Searching path array.
	jsonMap       *ymap.StrAnyMap  // The pared JSON objects for configuration files.
	violenceCheck bool             // Whether do violence check in value index searching. It affects the performance when set true(false in default).
}

const (
	DefaultName       = "config"             // DefaultName is the default group name for instance usage.
	DefaultConfigFile = "config.toml"        // DefaultConfigFile is the default configuration file name.
	cmdEnvKey         = "gf.gcfg"            // cmdEnvKey is the configuration key for command argument or environment.
	errorPrintKey     = "gf.gcfg.errorprint" // errorPrintKey is used to specify the key controlling error printing to stdout.
)

var (
	supportedFileTypes     = []string{"toml", "yaml", "yml", "json", "ini", "xml"}         // All supported file types suffixes.
	resourceTryFiles       = []string{"", "/", "config/", "config", "/config", "/config/"} // Prefix array for trying searching in resource manager.
	instances              = ymap.NewStrAnyMap(true)                                       // Instances map containing configuration instances.
	customConfigContentMap = ymap.NewStrStrMap(true)                                       // Customized configuration content.
)

// SetContent sets customized configuration content for specified `file`.
// The `file` is unnecessary param, default is DefaultConfigFile.
func SetContent(content string, file ...string) {
	name := DefaultConfigFile
	if len(file) > 0 {
		name = file[0]
	}
	// Clear file cache for instances which cached `name`.
	instances.LockFunc(func(m map[string]interface{}) {
		if customConfigContentMap.Contains(name) {
			for _, v := range m {
				v.(*Config).jsonMap.Remove(name)
			}
		}
		customConfigContentMap.Set(name, content)
	})
}

// GetContent returns customized configuration content for specified `file`.
// The `file` is unnecessary param, default is DefaultConfigFile.
func GetContent(file ...string) string {
	name := DefaultConfigFile
	if len(file) > 0 {
		name = file[0]
	}
	return customConfigContentMap.Get(name)
}

// RemoveContent removes the global configuration with specified `file`.
// If `name` is not passed, it removes configuration of the default group name.
func RemoveContent(file ...string) {
	name := DefaultConfigFile
	if len(file) > 0 {
		name = file[0]
	}
	// Clear file cache for instances which cached `name`.
	instances.LockFunc(func(m map[string]interface{}) {
		if customConfigContentMap.Contains(name) {
			for _, v := range m {
				v.(*Config).jsonMap.Remove(name)
			}
			customConfigContentMap.Remove(name)
		}
	})

	intlog.Printf(`RemoveContent: %s`, name)
}

// ClearContent removes all global configuration contents.
func ClearContent() {
	customConfigContentMap.Clear()
	// Clear cache for all instances.
	instances.LockFunc(func(m map[string]interface{}) {
		for _, v := range m {
			v.(*Config).jsonMap.Clear()
		}
	})

	intlog.Print(`RemoveConfig`)
}

// errorPrint checks whether printing error to stdout.
func errorPrint() bool {
	return ycmd.GetOptWithEnv(errorPrintKey, true).Bool()
}
