package tests

import (
	"macos-deployment/deploy-files/logger"
	"os"
	"testing"
)

type testLogger struct {
	Log              *logger.Log
	ProjectDirectory string
}

func GetLogger(t *testing.T) *testLogger {
	tempDir := t.TempDir()
	serialTag := "SERIAL_TAG"
	verbose := false

	logger := logger.NewLog(serialTag, tempDir+"/logs", verbose)

	testLogger := testLogger{
		Log:              logger,
		ProjectDirectory: tempDir,
	}

	return &testLogger
}

// Mkdir creates all directories in the temporary directory.
// The temporary directory appends itself to the given path by default.
func (t *testLogger) Mkdir(path string) error {
	err := os.MkdirAll(t.ProjectDirectory+"/"+path, 0o777)
	if err != nil {
		return err
	}

	return nil
}

// WriteFile writes a file at the given path in the temporary directory.
// The temporary directory appends itself to the given path by default.
//
// It does not create the parent directories.
func (t *testLogger) WriteFile(path string, data []byte) error {
	err := os.WriteFile(t.ProjectDirectory+"/"+path, data, 0o644)
	if err != nil {
		return err
	}
	return nil
}

// CheckError checks if it it is an error, then run t.Fatal.
func CheckError(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}
