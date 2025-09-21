package main

import (
	_ "embed"
	"fmt"
	embedhandler "macos-deployment/config"
	"macos-deployment/deploy-files/cmd"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"macos-deployment/deploy-files/utils"
	"macos-deployment/deploy-files/yaml"
	"os"
)

func main() {
	const projectName string = "macos-deployment"
	const distDirectory string = "dist"
	const zipFile string = "deploy.zip"

	config, err := yaml.NewConfig(embedhandler.YAMLBytes)
	if err != nil {
		// TODO: make this a better error message (incorrect keys, required keys missing, etc)
		fmt.Println("Error parsing YAML configuration")
		os.Exit(1)
	}

	// by default we will put in the tmp directory if none is given
	logDirectory := config.LogDirectory
	if logDirectory == "" {
		logDirectory = "/tmp"
	}

	// not exiting, just in case mac fails somehow. but there are checks for non-mac devices.
	serialTag, err := utils.GetSerialTag()
	log := logger.NewLog(serialTag, logDirectory)
	if err != nil {
		log.Error.Println("Unable to get serial number: %v", err)
	}

	metadata := utils.NewMetadata(projectName, serialTag, distDirectory, zipFile)
	scripts := scripts.NewScript()

	cmd.InitializeRoot(log, config, scripts, metadata)
	cmd.Execute()
}
