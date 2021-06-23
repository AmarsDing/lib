package ylog

import "bytes"

// Write implements the io.Writer interface.
// It just prints the content using Print.
func (l *Logger) Write(p []byte) (n int, err error) {
	l.Header(false).Print(string(bytes.TrimRight(p, "\r\n")))
	return len(p), nil
}
