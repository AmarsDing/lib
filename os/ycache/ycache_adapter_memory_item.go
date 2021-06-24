package ycache

import (
	"github.com/AmarsDing/lib/os/ytime"
)

// IsExpired checks whether <item> is expired.
func (item *adapterMemoryItem) IsExpired() bool {
	// Note that it should use greater than or equal judgement here
	// imagining that the cache time is only 1 millisecond.
	if item.e >= ytime.TimestampMilli() {
		return false
	}
	return true
}
