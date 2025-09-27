package main

import (
	_ "embed"
	"fmt"
	"macos-deployment/deploy-files/cmd"
	"macos-deployment/deploy-files/yaml"
	"os"
)

func main() {
	config := &yaml.Config{}

	err := config.SetAdminUsername()
	if err != nil {
		fmt.Println(err)
	}

	err = config.SetAdminPassword()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(1)
	cmd.InitializeRoot()
	cmd.Execute()
}
