package tests

import (
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/utils"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testLogger struct {
	Log              *logger.Log
	ProjectDirectory string
}

var filePerms = utils.Perms{
	Executable: 0o744,
	Full:       0o777,
	Base:       0o644,
	BaseDir:    0o755,
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
// The path appends itself to the temporary dictionary.
//
// It will create parent directories, but the base name must be a file.
func (t *testLogger) MkFile(path string, data []byte, perm os.FileMode) error {
	if strings.Contains(path, "/") {
		filename := filepath.Base(path)
		err := t.Mkdir(strings.ReplaceAll(path, "/"+filename, ""))
		if err != nil {
			return err
		}
	}

	err := os.WriteFile(t.ProjectDirectory+"/"+path, data, perm)
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
