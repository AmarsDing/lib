// Package ysha1 provides useful API for SHA1 encryption algorithms.
package ysha1

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"

	"github.com/AmarsDing/lib/util/yconv"
)

// Encrypt encrypts any type of variable using SHA1 algorithms.
// It uses yconv package to convert <v> to its bytes type.
func Encrypt(v interface{}) string {
	r := sha1.Sum(yconv.Bytes(v))
	return hex.EncodeToString(r[:])
}

// EncryptFile encrypts file content of <path> using SHA1 algorithms.
func EncryptFile(path string) (encrypt string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// MustEncryptFile encrypts file content of <path> using SHA1 algorithms.
// It panics if any error occurs.
func MustEncryptFile(path string) string {
	result, err := EncryptFile(path)
	if err != nil {
		panic(err)
	}
	return result
}
