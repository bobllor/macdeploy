package utils

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/bobllor/macdeploy/src/deploy-files/utils"
	"github.com/bobllor/macdeploy/src/tests"
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
	rootDir := t.TempDir()
	fileNames := []string{
		"test1.txt", "test2.txt",
		"test3.go", "test4.py",
		"test5.py", "test6.py",
	}

	// this does not get removed here
	pyDir := rootDir + "/python"

	err := os.MkdirAll(pyDir, 0o777)
	tests.Checkf(t, err != nil, "failed to create dir %s: %v", pyDir, err)

	pathsToRemove := []string{}

	for _, file := range fileNames {
		path := rootDir + "/" + file
		if strings.HasSuffix(file, ".py") {
			path = pyDir + "/" + file
		}

		err = os.WriteFile(path, []byte{}, 0o666)
		tests.Checkf(t, err != nil, "failed to write file %s: %v", path, err)

		pathsToRemove = append(pathsToRemove, path)
	}

	utils.RemoveFiles(pathsToRemove)

	for _, path := range pathsToRemove {
		_, err := os.Stat(path)
		tests.Checkf(t, err == nil, "failed to remove %s, file exists", path)
	}

	_, err = os.Stat(pyDir)
	tests.Checkf(t, err != nil, "failed to stat %s: %v", pyDir, err)
}

func TestRemoveFilesDir(t *testing.T) {
	rootDir := t.TempDir()
	fileNames := []string{
		"test1.txt", "test2.txt",
		"test3.go", "test4.py",
		"test5.py", "test6.py",
	}

	// this is removed
	pyDir := rootDir + "/python"

	err := os.MkdirAll(pyDir, 0o777)
	tests.Checkf(t, err != nil, "failed to create dir %s: %v", pyDir, err)

	pathsToRemove := []string{pyDir}

	for _, file := range fileNames {
		path := rootDir + "/" + file
		isPyFile := strings.HasSuffix(file, ".py")
		if isPyFile {
			path = pyDir + "/" + file
		}

		err = os.WriteFile(path, []byte{}, 0o666)
		tests.Checkf(t, err != nil, "failed to write file %s: %v", path, err)

		if !isPyFile {
			pathsToRemove = append(pathsToRemove, path)
		}
	}

	utils.RemoveFiles(pathsToRemove)

	files, err := os.ReadDir(rootDir)
	tests.Checkf(t, err != nil, "failed to read directory %s: %v", rootDir, err)

	tests.Checkf(t, len(files) != 0, "failed to remove files '%s'", strings.Join(pathsToRemove, ","))
}

func TestMetaToString(t *testing.T) {
	root := t.TempDir()
	zipName := "deploy.zip"
	st := "12345"

	meta := utils.NewMetadata(st, root, root+"/"+zipName)

	baseData := []string{
		root,
		zipName,
		st,
	}

	str := meta.ToString()

	for _, v := range baseData {
		tests.Checkf(t, strings.Contains(str, v) == false, "Failed to find %s in meta string %s", v, str)
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
