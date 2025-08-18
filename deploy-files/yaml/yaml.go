package yaml

import (
	"errors"
	"fmt"
	"macos-deployment/deploy-files/logger"
	"os"

	"github.com/goccy/go-yaml"
)

// ReadYAML reads the YAML configuration file and returns a struct of the file.
//
// If there is an issue with reading the YAML configuration then this will exit the script.
func ReadYAML(configPath string) *Config {
	yamlConfig := &Config{}

	content, err := os.ReadFile(configPath)
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading YAML config: %s", err.Error()), 2)
		panic(err)
	}

	err = yaml.Unmarshal(content, yamlConfig)
	if err != nil {
		logger.Log(fmt.Sprintf("Error parsing YAML config: %s", err.Error()), 2)
		panic(err)
	}

	err = validateYAML(yamlConfig)
	if err != nil {
		logger.Log(fmt.Sprintf("YAML is missing required values: %s", err.Error()), 2)
		panic(err)
	}

	return yamlConfig
}

// ValidateYAML checks for missing required values in the YAML config.
//
// The only required value is the Admin.
func validateYAML(yamlConfig *Config) error {
	newError := func(msg string) error {
		return errors.New(msg)
	}

	if yamlConfig.Admin.User_Name == "" {
		err := newError("missing admin username or it cannot be empty")
		return err
	}

	if yamlConfig.Admin.Password == "" {
		err := newError("missing admin password or it cannot be empty")
		return err
	}

	return nil
}
