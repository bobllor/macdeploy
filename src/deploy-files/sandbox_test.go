package sandbox

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

func TestMain(t *testing.T) {
	testFileMap := map[string]struct{}{
		"removeme1.txt": {},
		"removeme2.py":  {},
		"somedir":       {},
	}
	fmt.Println(testFileMap)

	fmt.Println(os.Getwd())

	//utils.RemoveFiles(testFileMap)
	fmt.Println(runtime.GOARCH)
}
