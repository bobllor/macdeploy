package utils

import (
	"fmt"
	"macos-deployment/deploy-files/utils"
	"testing"
)

func TestName(t *testing.T) {
	namesList := []string{
		"John Doe", "John  Smith", "Jacob  B  Doe\n  ",
		"  Jessica \n\nThompson    lol", "someothernamehereOkay",
	}

	for _, name := range namesList {
		fmt.Println(utils.FormatFullName(name))
	}
}

/*
func TestGetFileMap(t *testing.T) {
	// contains a correct and incorrect path
	paths := [2]string{Home, "/non-existent/path"}
	for _, path := range paths {
		mapOut, mapErr := GetFileMap(path)
		if mapErr != nil {
			t.Errorf("Invalid path %p\n", mapErr)
		}
		fmt.Printf("Map: %v\n", mapOut)
	}
}
*/
