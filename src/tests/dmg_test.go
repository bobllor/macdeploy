package tests

import (
	"macos-deployment/deploy-files/core"
	"macos-deployment/deploy-files/scripts"
	"os"
	"testing"
)

var testDmgs = []string{
	"test.dmg", "another one.dmg",
}

var baseLen int = len(testDmgs)

func TestReadDmg(t *testing.T) {
	dir := t.TempDir()

	for _, dmgFile := range testDmgs {
		err := os.WriteFile(dir+"/"+dmgFile, []byte{}, 0o744)
		// hmm...
		if err != nil {
			t.Error(err)
		}
	}

	logger := GetLogger(dir + "/logs")
	scripts := scripts.NewScript()

	dmg := core.NewDmg(logger, scripts)

	dmgFiles, err := dmg.ReadDmgDirectory(dir)
	if err != nil {
		t.Errorf("failed to read directory: %v", err)
	}

	// there is an empty string added to the array
	newLen := len(dmgFiles) - 1

	if newLen != baseLen {
		t.Errorf("got %d, did not match the baseline %d", newLen, baseLen)
	}
}
