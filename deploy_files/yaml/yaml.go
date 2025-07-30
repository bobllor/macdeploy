package yaml

import (
	"macos-deployment/deploy_files/utils"
	"os"

	"github.com/goccy/go-yaml"
)

var yamlConfig utils.Config

// ReadYAML reads the YAML configuration file and returns a struct of the file.
func ReadYAML(configPath string) utils.Config {
	content, err := os.ReadFile(configPath)
	utils.CheckError(err)

	err = yaml.Unmarshal(content, &yamlConfig)
	utils.CheckError(err)

	return yamlConfig
}
