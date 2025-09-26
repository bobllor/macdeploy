package tests

import (
	"macos-deployment/deploy-files/logger"
	"testing"
)

type testLogger struct {
	Log           *logger.Log
	MainDirectory string
}

func GetLogger(t *testing.T) *testLogger {
	tempDir := t.TempDir()
	serialTag := "SERIAL_TAG"
	verbose := false

	logger := logger.NewLog(serialTag, tempDir+"/logs", verbose)

	testLogger := testLogger{
		Log:           logger,
		MainDirectory: tempDir,
	}

	return &testLogger
}
