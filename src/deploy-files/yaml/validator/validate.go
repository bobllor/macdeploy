package main

import (
	"fmt"
	embedhandler "macos-deployment/config"
	"macos-deployment/deploy-files/yaml"
	"os"
)

// main is used to read a config file from a given path and validate it.
// It will exit 1 if validation fails, or 0 if it succeeds.
func main() {
	config, err := yaml.NewConfig(embedhandler.YAMLBytes)
	if err != nil {
		fmt.Printf("Failed to read config file: %v\n", err)
		os.Exit(1)
	}

	err = yaml.Validate(config)
	if err != nil {
		fmt.Printf("Config file failed to validate %v\n", err)
		os.Exit(1)
	}
}
