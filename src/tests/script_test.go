package tests

import (
	_ "embed"
	"macos-deployment/deploy-files/scripts"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var script *scripts.BashScripts = scripts.NewScript()

func TestFindFiles(t *testing.T) {
	dir := t.TempDir()

	filesToWrite := []string{
		"sample_file.pkg", "teamviewer.pkg",
		"chrome.pkg", "some name.txt",
	}

	baseCount := 0

	for _, file := range filesToWrite {
		if strings.Contains(file, ".pkg") {
			baseCount += 1
		}
	}

	for _, fileName := range filesToWrite {
		err := os.WriteFile(dir+"/"+fileName, []byte{}, 0o644)
		if err != nil {
			t.Error(err)
		}
	}

	out, err := exec.Command("bash", "-c", script.FindFiles, dir, "*.pkg").Output()
	if err != nil {
		t.Error(err)
	}

	outArr := strings.Split(strings.TrimSpace(string(out)), "\n")

	if len(outArr) != baseCount {
		t.Errorf("failed to find files: got %d, expected %d", len(outArr), baseCount)
	}
}
