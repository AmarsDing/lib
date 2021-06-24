package ysession

import (
	"errors"

	"github.com/AmarsDing/lib/util/yuid"
)

var (
	ErrorDisabled = errors.New("this feature is disabled in this storage")
)

// NewSessionId creates and returns a new and unique session id string,
// which is in 36 bytes.
func NewSessionId() string {
	return yuid.S()
}
