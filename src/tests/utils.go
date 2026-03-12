package tests

import (
	"testing"
)

// Fatal checks if the error is not nil, if it is then
// FailNow with the given msg string. Otherwise, do
// nothing.
func Fatal(t *testing.T, err error, msg string) {
	if err != nil {
		t.Fatal(msg)
	}
}

// Checkf uses a FailNow with the format string as its message by
// checking if the status is true. If status is false, then do nothing.
func Checkf(t *testing.T, status bool, format string, v ...any) {
	if status {
		t.Fatalf(format, v...)
	}
}
