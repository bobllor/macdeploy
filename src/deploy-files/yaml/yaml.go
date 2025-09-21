package yaml

import (
	"errors"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Accounts          map[string]UserInfo `yaml:"accounts"`
	Packages          map[string][]string `yaml:"packages"`
	SearchDirectories []string            `yaml:"search_directories"`
	Admin             UserInfo
	ServerHost        string `yaml:"server_host"`
	FileVault         bool
	Firewall          bool
	LogDirectory      string `yaml:"log_directory"`
	AlwaysCleanup     bool   `yaml:"always_cleanup"`
}

type UserInfo struct {
	Username       string `yaml:"username"`
	Password       string
	IgnoreAdmin    bool `yaml:"ignore_admin"`
	ChangePassword bool `yaml:"change_password"`
}

// NewConfig returns a struct containing data read from the YAML file. The file is read
// through embedding.
//
// If an issue occurs while reading the file then it will return an error.
func NewConfig(data []byte) (*Config, error) {
	config := Config{}

	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	err = config.validateYAML()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// ValidateYAML checks for missing required values in the YAML config.
//
// The only required value is the Admin.
func (u *Config) validateYAML() error {
	newError := func(msg string) error {
		return errors.New(msg)
	}

	if u.Admin.Username == "" {
		err := newError("missing admin username, it cannot be empty")
		return err
	}

	if u.Admin.Password == "" {
		err := newError("missing admin password, it cannot be empty")
		return err
	}

	return nil
}
