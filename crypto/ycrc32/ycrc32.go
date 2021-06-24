// Package ycrc32 provides useful API for CRC32 encryption algorithms.
package ycrc32

import (
	"hash/crc32"

	"github.com/AmarsDing/lib/util/yconv"
)

// Encrypt encrypts any type of variable using CRC32 algorithms.
// It uses gconv package to convert <v> to its bytes type.
func Encrypt(v interface{}) uint32 {
	return crc32.ChecksumIEEE(yconv.Bytes(v))
}
