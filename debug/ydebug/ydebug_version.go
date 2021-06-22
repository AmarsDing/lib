package ydebug

import (
	"io/ioutil"
	"strconv"

	"github.com/AmarsDing/lib/crypto/ymd5"
	"github.com/AmarsDing/lib/encoding/yhash"
)

// BinVersion returns the version of current running binary.
// It uses yhash.BKDRHash+BASE36 algorithm to calculate the unique version of the binary.
func BinVersion() string {
	if binaryVersion == "" {
		binaryContent, _ := ioutil.ReadFile(selfPath)
		binaryVersion = strconv.FormatInt(
			int64(yhash.BKDRHash(binaryContent)),
			36,
		)
	}
	return binaryVersion
}

// BinVersionMd5 returns the version of current running binary.
// It uses MD5 algorithm to calculate the unique version of the binary.
func BinVersionMd5() string {
	if binaryVersionMd5 == "" {
		binaryVersionMd5, _ = ymd5.EncryptFile(selfPath)
	}
	return binaryVersionMd5
}
