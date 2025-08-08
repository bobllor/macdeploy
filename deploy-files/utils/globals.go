package utils

import (
	"os"
	"runtime"
	"strings"
)

// all client-side files and directories will be placed in the home directory

var Home string = os.Getenv("HOME")
var ProjectDir string = Home + "/macos-deployment"

var MainDir string

var ScriptDir string = "scripts"
var PKGPath string = Home + "/pkg-files"

var SerialTag string

func init() {
	_, file, _, _ := runtime.Caller(0)

	paths := strings.Split(file, "/")

	var mainDirIndex int

	for i, path := range paths {
		if strings.Contains(strings.ToLower(path), "macos-deployment") {
			mainDirIndex = i
			break
		}
	}

	mainPath := strings.Join(paths[:mainDirIndex+1], "/")

	MainDir = mainPath
}
