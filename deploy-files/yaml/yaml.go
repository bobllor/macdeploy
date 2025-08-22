package yaml

import (
	"errors"
	"fmt"
	"macos-deployment/deploy-files/logger"

	"github.com/goccy/go-yaml"
)

// ReadYAML reads the YAML configuration file and returns a struct of the file.
//
// If there is an issue with reading the YAML configuration then this will exit the script.
func ReadYAML(data []byte) *Config {
	yamlConfig := &Config{}

	err := yaml.Unmarshal(data, yamlConfig)
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
