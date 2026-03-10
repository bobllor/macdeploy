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
