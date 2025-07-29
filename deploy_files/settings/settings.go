package settings

import (
	"macos-deployment/deploy_files/utils"
	"os"

	"github.com/goccy/go-yaml"
)

var settings struct {
	Accounts           map[string]map[string]string
	Packages           []string
	Search_Directories []string
}

type Settings struct {
	Accounts           map[string]map[string]string
	Packages           []string
	Search_Directories []string
}

// ReadYAML returns a struct settings containing the key-values of the YAML for the script.
func ReadYAML(configPath string) Settings {
	content, err := os.ReadFile(configPath)
	utils.CheckError(err)

	err = yaml.Unmarshal(content, &settings)
	utils.CheckError(err)

	return settings
}
