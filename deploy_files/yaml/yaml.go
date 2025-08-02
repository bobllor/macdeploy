package yaml

import (
	"macos-deployment/deploy_files/utils"
	"os"

	"github.com/goccy/go-yaml"
)

var yamlConfig utils.Config

// ReadYAML reads the YAML configuration file and returns a struct of the file.
//
// If there is an issue with reading the YAML configuration then this will exit the script.
func ReadYAML(configPath string) utils.Config {
	content, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(content, &yamlConfig)
	if err != nil {
		panic(err)
	}

	return yamlConfig
}
