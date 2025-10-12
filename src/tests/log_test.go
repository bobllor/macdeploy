package tests

import (
	"macos-deployment/deploy-files/logger"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestDirFileNames(t *testing.T) {
	log := GetLogger(t)

	if !strings.Contains(log.Log.GetLogName(), ".log") {
		t.Fatalf("log file name failed to generate: %s", log.Log.GetLogName())
	}

	if !strings.Contains(log.Log.GetLogPath(), log.ProjectDirectory) {
		t.Fatalf("generated log path %s does not match base path %s", log.Log.GetLogPath(), log.ProjectDirectory)
	}
}

func TestMkDir(t *testing.T) {
	log := GetLogger(t)

	err := logger.MkdirAll(log.Log.GetLogPath(), 0o744)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(log.ProjectDirectory)
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
		baseDirName + "\\/////",
		tempDir + "\\this\\is\\a\\dir",
		baseDirName + "/a.log/b.log",
		baseDirName + "/a.log/b.log/c.log/d.log//",
	}

	for _, dir := range dirs {
		newDir := logger.FormatLogPath(dir)

		if newDir != baseDirName {
			t.Fatalf("formatting directory %s failed: got %s", dir, newDir)
		}
	}
}

func TestLogHomeExansionFormat(t *testing.T) {
	baseline := []string{
		"./~test", os.Getenv("HOME"),
		os.Getenv("HOME") + "/~", "/~",
		"./~~",
	}

	logPaths := []string{
		"~test", "~",
		"/~", "~~",
		"////~", "~/~//",
	}

	for _, path := range logPaths {
		newPath := logger.FormatLogPath(path)

		if !slices.Contains(baseline, newPath) {
			t.Fatalf("path: %s, got %s, failed to meet baseline output", path, newPath)
		}
	}
}

func TestLogSpecialFormat(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	baseline := []string{
		os.Getenv("HOME") + "/logs",
		wd, "./...",
	}

	logPaths := []string{
		".", "/", "./", "",
		"///", "\\////", "...",
	}

	for _, path := range logPaths {
		newPath := logger.FormatLogPath(path)

		if !slices.Contains(baseline, newPath) {
			t.Fatalf("path: %s, got %s, failed to meet baseline output", path, newPath)
		}
	}
}

func TestLogf(t *testing.T) {
	log := GetLogger(t)

	err := logger.MkdirAll(log.Log.GetLogDirectory(), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	baseMsg := "This is a test message"
	fmtMsg := "This is a test %s"

	log.Log.Info.Logf(fmtMsg, "message")

	err = log.Log.WriteFile()
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(log.Log.GetLogPath())
	if err != nil {
		t.Fatal(err)
	}

	contentStr := strings.TrimSpace(string(content))

	if !strings.Contains(contentStr, baseMsg) {
		t.Fatalf(`Did not get baseline "%s" in file, got: "%s"`, baseMsg, contentStr)
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
