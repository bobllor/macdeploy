package utils

import (
	"os"
	"runtime"
	"strings"
)

// all client-side files and directories will be placed in the home directory

var Globals = &Global{}

// InitializeGlobals initializes the global variables for paths and device information.
func InitializeGlobals() {
	Globals.ZIPFileName = "deploy.zip"
	Globals.ARMBinaryName = "deploy-arm.bin"
	Globals.X86_64BinaryName = "deploy-x86_64.bin"
	Globals.DistDirName = "dist"
	Globals.ProjectName = "macos-deployment"

	initPaths()
	initSerialTag()
}

func initPaths() {
	Globals.Home = os.Getenv("HOME")
	Globals.ProjectPath = getProjectPath()
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
		if strings.Contains(strings.ToLower(path), Globals.ProjectName) {
			mainDirIndex = i
			break
		}
	}

	mainPath := strings.Join(paths[:mainDirIndex+1], "/")

	return mainPath
}
