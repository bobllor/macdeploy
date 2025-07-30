package utils

import "os"

// all client-side files and directories will be placed in the home directory

var Home string = os.Getenv("HOME")

var ConfigPath string = "./config.yaml"
var ScriptDir string = "./client-files"
var PKGPath string = Home + "/Pictures" // FIXME: remove later for prod
