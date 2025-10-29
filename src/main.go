package main

import (
	_ "embed"
	"macos-deployment/deploy-files/cmd"
)

func main() {
	cmd.InitializeRoot()
	cmd.InitializeUserCmd()
	cmd.Execute()
}
