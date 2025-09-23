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

	// not exiting, just in case mac fails somehow. but there are checks for non-mac devices.
	serialTag, err := utils.GetSerialTag()
	if err != nil {
		serialTag = "UNKNOWN"
		fmt.Printf("Unable to get serial number: %v\n", err)
	}

	metadata := utils.NewMetadata(projectName, serialTag, distDirectory, zipFile)
	scripts := scripts.NewScript()
	config, err := yaml.NewConfig(embedhandler.YAMLBytes)
	if err != nil {
		// TODO: make this a better error message (incorrect keys, required keys missing, etc)
		fmt.Printf("Error parsing YAML configuration, %v\n", err)
		os.Exit(1)
	}

	// by default we will put in the home directory if none is given
	logDirectory := config.LogOutput
	defaultLogDir := fmt.Sprintf("%s/%s", metadata.Home, ".macdeploy")
	if logDirectory == "" {
		logDirectory = defaultLogDir
	} else {
		logDirectory = logger.FormatLogOutput(logDirectory)
	}

	err = logger.MkdirAll(logDirectory, 0o744)
	if err != nil {
		fmt.Printf("Unable to make logging directory: %v\n", err)
		fmt.Printf("Changing log output to home directory: %s\n", defaultLogDir)
		logDirectory = defaultLogDir

		err = logger.MkdirAll(defaultLogDir, 0o744)
		if err != nil {
			fmt.Printf("Unable to make logging directory: %v\n", err)
		}
	}

	log := logger.NewLog(serialTag, logDirectory)
	log.Info.Log("Log directory: %s", logDirectory)

	cmd.InitializeRoot(log, config, scripts, metadata)
	cmd.Execute()
}
