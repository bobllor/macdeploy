package settings

import (
	"macos-deployment/utils"
	"os"

	"github.com/goccy/go-yaml"
)

// runs in macos-deployment
var configPath string = "../config/settings.yaml"

var settings struct {
	Accounts map[string]map[string]string
	Packages []string
}

type Settings struct {
	Accounts map[string]map[string]string
	Packages []string
}

// ReadYAML returns a struct settings containing the key-values of the YAML for the script.
func ReadYAML() Settings {
	content, err := os.ReadFile(configPath)
	utils.CheckError(err)

	err = yaml.Unmarshal(content, &settings)
	utils.CheckError(err)

	return settings
}
