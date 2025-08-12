package yaml

import (
	"os"

	"github.com/goccy/go-yaml"
)

// ReadYAML reads the YAML configuration file and returns a struct of the file.
//
// If there is an issue with reading the YAML configuration then this will exit the script.
func ReadYAML(configPath string) Config {
	var yamlConfig Config

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
