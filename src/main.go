package main

import (
	_ "embed"

	"github.com/bobllor/macdeploy/src/deploy-files/cmd"
)

func main() {
	cmd.InitializeRoot()
	cmd.InitializeUserCmd()
	cmd.InitializeInstallCmd()

	cmd.Execute()
}
