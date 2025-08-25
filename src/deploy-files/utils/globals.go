package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// all client-side files and directories will be placed in the home directory

var Globals = &Global{}

// InitializeGlobals initializes the global variables for paths and device information.
func InitializeGlobals() {
	initPaths()
	initSerialTag()

	Globals.BinaryName = "deploy.bin"
	Globals.PKGDirName = "pkg-files"
	Globals.ZIPFileName = "deploy.zip"
	Globals.AMDBinaryName = "deploy-amd64.bin"
}

func initPaths() {
	Globals.Home = os.Getenv("HOME")
	Globals.ProjectPath = getProjectPath()

	pkgFileName := "pkg-files"
	Globals.PKGPath = fmt.Sprintf("%s/%s", Globals.Home, pkgFileName)
}

func initSerialTag() {
	serialTag, err := GetSerialTag()
	if err != nil {
		serialTag = "UNKNOWN"
	}

	Globals.SerialTag = serialTag
}

func getProjectPath() string {
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

	return mainPath
}
