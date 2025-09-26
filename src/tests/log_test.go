package tests

import (
	"macos-deployment/deploy-files/logger"
	"os"
	"strings"
	"testing"
)

func TestDirFileNames(t *testing.T) {
	log := GetLogger(t)

	if !strings.Contains(log.Log.GetLogName(), ".log") {
		t.Fatalf("log file name failed to generate: %s", log.Log.GetLogName())
	}

	if !strings.Contains(log.Log.GetLogPath(), log.MainDirectory) {
		t.Fatalf("generated log path %s does not match base path %s", log.Log.GetLogPath(), log.MainDirectory)
	}
}

func TestMkDir(t *testing.T) {
	log := GetLogger(t)

	err := logger.MkdirAll(log.Log.GetLogPath(), 0o744)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(log.MainDirectory)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFormatDir(t *testing.T) {
	tempDir := t.TempDir()
	baseDirName := tempDir + "/this/is/a/dir"

	dirs := []string{
		baseDirName + "////",
		baseDirName + "/",
		baseDirName + "/a-name.log",
		baseDirName + "\\",
		tempDir + "\\this\\is\\a\\dir",
		baseDirName,
		baseDirName + "/a.log/b.log",
		baseDirName + "/a.log/b.log/c.log/d.log//",
	}

	for _, dir := range dirs {
		newDir := logger.FormatLogOutput(dir)

		if newDir != baseDirName {
			t.Fatalf("formatting directory %s failed: got %s", dir, newDir)
		}
	}

	singleDirTests := []string{"/", "\\", "///"}
	for _, dir := range singleDirTests {
		newDir := logger.FormatLogOutput(dir)

		if newDir != "." {
			t.Fatalf("formatting directory %s failed: got %s", dir, newDir)
		}
	}
}

func TestWriteLog(t *testing.T) {
	log := GetLogger(t)

	err := logger.MkdirAll(log.Log.GetLogDirectory(), 0o744)
	if err != nil {
		t.Fatal(err)
	}

	errorMsg := "AN ERROR HERE"
	warn := "A WARNING HERE"
	info := "A INFO HERE"
	debug := "A DEBUG HERE"

	log.Log.Error.Log(errorMsg)
	log.Log.Warn.Log(warn)
	log.Log.Info.Log(info)
	log.Log.Debug.Log(debug)

	err = log.Log.WriteFile()
	if err != nil {
		t.Fatal(err)
	}

	contentBytes, err := os.ReadFile(log.Log.GetLogPath())
	if err != nil {
		t.Fatal(err)
	}

	messages := []string{errorMsg, warn, info, debug}
	content := string(contentBytes)

	for _, msg := range messages {
		if !strings.Contains(content, msg) {
			t.Fatalf("%s could not be found in %s", msg, content)
		}
	}
}
