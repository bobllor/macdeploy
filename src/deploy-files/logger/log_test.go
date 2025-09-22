package logger

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestDirFileNames(t *testing.T) {
	dir := GetDir(t, "cool/path/bro")
	log := GetLog(t, dir)

	if !strings.Contains(log.logFileName, ".log") {
		t.Fatalf("log file name failed to generate: %s", log.logFileName)
	}

	if !strings.Contains(log.logFilePath, dir) {
		t.Fatalf("generated log path %s does not match base path %s", log.logFilePath, dir)
	}
}

func TestMkDir(t *testing.T) {
	dir := GetDir(t, "some/dir/here")
	log := GetLog(t, dir)

	err := MkdirAll(log.logFilePath, 0o744)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(dir)
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
	}

	for _, dir := range dirs {
		newDir := FormatLogOutput(dir)

		if newDir != baseDirName {
			t.Fatalf("formatting directory %s failed: got %s", dir, newDir)
		}
	}

	// this takes advantage of the fact that splitting these adds an empty string to the array
	// if google changes it then i have to fix this lol
	singleDirTests := []string{"/", "\\"}
	for _, dir := range singleDirTests {
		newDir := FormatLogOutput(dir)

		if newDir != "/" {
			t.Fatalf("formatting directory %s failed: got %s", dir, newDir)
		}
	}
}

func TestWriteLog(t *testing.T) {
	dir := GetDir(t, "some/dir")
	log := GetLog(t, dir)

	err := MkdirAll(dir, 0o744)
	if err != nil {
		t.Fatal(err)
	}

	errorMsg := "AN ERROR HERE"
	warn := "A WARNING HERE"
	info := "A INFO HERE"
	debug := "A DEBUG HERE"

	log.Error.Log(errorMsg)
	log.Warn.Log(warn)
	log.Info.Log(info)
	log.Debug.Log(debug)

	err = log.WriteFile()
	if err != nil {
		t.Fatal(err)
	}

	contentBytes, err := os.ReadFile(log.logFilePath)
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

func GetLog(t *testing.T, dirPath string) *Log {
	serialTag := "LOL12345"

	return NewLog(serialTag, dirPath)
}

func GetDir(t *testing.T, dirName string) string {
	return fmt.Sprintf("%s/%s", t.TempDir(), dirName)
}
