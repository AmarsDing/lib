package yredis

import (
	"github.com/AmarsDing/lib/errors/yerror"
	"github.com/AmarsDing/lib/internal/intlog"

	"github.com/AmarsDing/lib/container/ymap"
	"github.com/AmarsDing/lib/text/yregex"
	"github.com/AmarsDing/lib/text/ystr"
	"github.com/AmarsDing/lib/util/yconv"
)

const (
	DefaultGroupName = "default" // Default configuration group name.
	DefaultRedisPort = 6379      // Default redis port configuration if not passed.
)

var (
	// Configuration groups.
	configs = ymap.NewStrAnyMap(true)
)

// SetConfig sets the global configuration for specified group.
// If <name> is not passed, it sets configuration for the default group name.
func SetConfig(config *Config, name ...string) {
	group := DefaultGroupName
	if len(name) > 0 {
		group = name[0]
	}
	configs.Set(group, config)
	instances.Remove(group)

	intlog.Printf(`SetConfig for group "%s": %+v`, group, config)
}

// SetConfigByStr sets the global configuration for specified group with string.
// If <name> is not passed, it sets configuration for the default group name.
func SetConfigByStr(str string, name ...string) error {
	group := DefaultGroupName
	if len(name) > 0 {
		group = name[0]
	}
	config, err := ConfigFromStr(str)
	if err != nil {
		return err
	}
	configs.Set(group, config)
	instances.Remove(group)
	return nil
}

// GetConfig returns the global configuration with specified group name.
// If <name> is not passed, it returns configuration of the default group name.
func GetConfig(name ...string) (config *Config, ok bool) {
	group := DefaultGroupName
	if len(name) > 0 {
		group = name[0]
	}
	if v := configs.Get(group); v != nil {
		return v.(*Config), true
	}
	return &Config{}, false
}

// RemoveConfig removes the global configuration with specified group.
// If <name> is not passed, it removes configuration of the default group name.
func RemoveConfig(name ...string) {
	group := DefaultGroupName
	if len(name) > 0 {
		group = name[0]
	}
	configs.Remove(group)
	instances.Remove(group)

	intlog.Printf(`RemoveConfig: %s`, group)
}

// ConfigFromStr parses and returns config from given str.
// Eg: host:port[,db,pass?maxIdle=x&maxActive=x&idleTimeout=x&maxConnLifetime=x]
func ConfigFromStr(str string) (config *Config, err error) {
	array, _ := yregex.MatchString(`^([^:]+):*(\d*),{0,1}(\d*),{0,1}(.*)\?(.+)$`, str)
	if len(array) == 6 {
		parse, _ := ystr.Parse(array[5])
		config = &Config{
			Host: array[1],
			Port: yconv.Int(array[2]),
			Db:   yconv.Int(array[3]),
			Pass: array[4],
		}
		if config.Port == 0 {
			config.Port = DefaultRedisPort
		}
		if err = yconv.Struct(parse, config); err != nil {
			return nil, err
		}
		return
	}
	array, _ = yregex.MatchString(`([^:]+):*(\d*),{0,1}(\d*),{0,1}(.*)`, str)
	if len(array) == 5 {
		config = &Config{
			Host: array[1],
			Port: yconv.Int(array[2]),
			Db:   yconv.Int(array[3]),
			Pass: array[4],
		}
		if config.Port == 0 {
			config.Port = DefaultRedisPort
		}
	} else {
		err = yerror.Newf(`invalid redis configuration: "%s"`, str)
	}
	return
}

// ClearConfig removes all configurations and instances of redis.
func ClearConfig() {
	configs.Clear()
	instances.Clear()
}
