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
	log := GetLogger(t)

	for _, dmgFile := range testDmgs {
		err := os.WriteFile(log.MainDirectory+"/"+dmgFile, []byte{}, 0o744)
		// hmm...
		if err != nil {
			t.Error(err)
		}
	}

	scripts := scripts.NewScript()

	dmg := core.NewDmg(log.Log, scripts)

	dmgFiles, err := dmg.ReadDmgDirectory(log.MainDirectory)
	if err != nil {
		t.Errorf("failed to read directory: %v", err)
	}

	// there is an empty string added to the array
	newLen := len(dmgFiles)

	if newLen != baseLen {
		t.Errorf("got %d, did not match the baseline %d", newLen, baseLen)
	}
}
