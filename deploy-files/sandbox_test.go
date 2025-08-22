package sandbox

import (
	"macos-deployment/deploy-files/utils"
	"testing"
)

func TestMain(t *testing.T) {
	testFileMap := map[string]struct{}{
		"removeme1.txt": {},
		"removeme2.py":  {},
		"somedir":       {},
	}

	utils.RemoveFiles(testFileMap)
}
