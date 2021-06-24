package yview

import "github.com/AmarsDing/lib/container/ymap"

const (
	// Default group name for instance usage.
	DefaultName = "default"
)

var (
	// Instances map.
	instances = ymap.NewStrAnyMap(true)
)

// Instance returns an instance of View with default settings.
// The parameter <name> is the name for the instance.
func Instance(name ...string) *View {
	key := DefaultName
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}
	return instances.GetOrSetFuncLock(key, func() interface{} {
		return New()
	}).(*View)
}
