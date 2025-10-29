package tests

import (
	"fmt"
	"macos-deployment/deploy-files/utils"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestGetFiles(t *testing.T) {
	// this does not search recursively.
	fakeFiles := []string{
		"test1.txt", "test2.txt",
		"test3.go", "test4.py",
	}

	tempDir := t.TempDir()

	for _, file := range fakeFiles {
		_, err := os.Create(tempDir + "/" + file)
		if err != nil {
			t.Fatalf("Failed to create files: %v", err)
		}
	}

	files, err := utils.GetFiles(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	slices.Sort(fakeFiles)
	slices.Sort(files)

	for i, file := range files {
		baseFile := fakeFiles[i]

		if file != baseFile {
			t.Fatalf("Expected %s got %s", baseFile, file)
		}
	}
}

func TestRemoveFiles(t *testing.T) {
	fakeFiles := []string{
		"test1.txt", "test2.txt",
		"test3.go", "test4.py",
	}

	tempDir := t.TempDir()
	filesToRemove := map[string]any{}

	for _, file := range fakeFiles {
		_, err := os.Create(tempDir + "/" + file)
		if err != nil {
			t.Fatalf("Failed to create files: %v", err)
		}
		filesToRemove[file] = false
	}

	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal("Failed to read directory")
	}

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal("Failed to change directories")
	}

	utils.RemoveFiles(filesToRemove, files)

	files, err = os.ReadDir(tempDir)
	if err != nil {
		t.Fatal("Failed to read directory")
	}

	if len(files) != 0 {
		t.Fatal("Failed to remove files")
	}
}

func TestFormatImportantString(t *testing.T) {
	baseLines := []string{
		"An example",
		"Message goes",
		"Here, with padding of two",
		"Another message goes here but longer than the other three",
		fmt.Sprintf("String format test %d", 1),
	}

	msg := utils.FormatBannerString(baseLines, 2)

	msgLines := strings.Split(msg, "\n")
	msgLines = msgLines[:len(msgLines)-2]

	baseLineIndex := 0
	for i := 1; i < len(msgLines); i++ {
		line := strings.TrimSpace(msgLines[i])
		if line == "" {
			continue
		}

		baseLine := baseLines[baseLineIndex]

		if line != baseLine {
			t.Fatalf("Got %s expected %s", line, baseLine)
		}
		baseLineIndex += 1
	}
}
