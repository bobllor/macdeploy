package utils

import (
	"os"
)

// all client-side files and directories will be placed in the home directory

var Home string = os.Getenv("HOME")
var MainDir string = Home + "/macos-deployment"

var ScriptDir string = "deploy_files"
var PKGPath string = Home + "pkg-files" // FIXME: for prod this is going to be "Home/pkg-files"

var SerialTag string
