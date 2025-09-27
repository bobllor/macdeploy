package yaml

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/goccy/go-yaml"
	"golang.org/x/term"
)

type Config struct {
	Accounts          map[string]UserInfo `yaml:"accounts"`
	Packages          map[string][]string `yaml:"packages"`
	SearchDirectories []string            `yaml:"search_directories"`
	Admin             UserInfo
	ServerHost        string `yaml:"server_host"`
	FileVault         bool
	Firewall          bool
	LogOutput         string `yaml:"log_output"`
}

type UserInfo struct {
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	IgnoreAdmin    bool   `yaml:"ignore_admin"`
	ChangePassword bool   `yaml:"change_password"`
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

	return &config, nil
}

// SetUsername is used to set the username if one was not given.
// This uses a command execution with whoami.
//
// It returns an error if the command fails to run.
func (u *UserInfo) SetUsername() error {
	// prevents accidental runs
	if u.Username != "" {
		return nil
	}

	out, err := exec.Command("whoami").Output()
	if err != nil {
		return err
	}

	user := strings.TrimSpace(string(out))
	u.Username = user

	return nil
}

// SetPassword is used to set the password of the user if one was not given.
// It prompts a hidden input for the user password.
//
// It returns an error if the maximum attempt is reached or if an error occurs.
// By default the maximum attempts is 3.
func (u *UserInfo) SetPassword() error {
	fmt.Print("Enter the password: ")
	pwOne, err := u.readPassword()
	if err != nil {
		return err
	}
	if pwOne == "" {
		return errors.New("cannot have empty password")
	}

	maxAttempts := 3
	attempts := 0

	for attempts < maxAttempts {
		fmt.Print("Enter the password again: ")
		pwTwo, err := u.readPassword()
		if err != nil {
			return err
		}

		if pwTwo == pwOne {
			break
		}

		fmt.Println("Sorry, try again")
		attempts += 1
	}

	if attempts >= maxAttempts {
		return fmt.Errorf("%d incorrect password attempts", attempts)
	}

	u.Password = pwOne

	return nil
}

// readPassword reads the input from STDIN securely.
func (u *UserInfo) readPassword() (string, error) {
	stdin := int(syscall.Stdin)

	oldState, err := term.GetState(stdin)
	if err != nil {
		return "", err
	}
	defer term.Restore(stdin, oldState)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		for _ = range ch {
			term.Restore(stdin, oldState)
			os.Exit(1)
		}
	}()

	pwBytes, err := term.ReadPassword(stdin)
	if err != nil {
		return "", err
	}

	// formatting, yes...
	fmt.Println()

	return strings.TrimSpace(string(pwBytes)), nil
}

// InitializeSudo starts a sudo session without the need of manual input.
// This can be called multiple times to refresh the sudo timer.
func (u *UserInfo) InitializeSudo() error {
	initSudoCmd := fmt.Sprintf("sudo -S echo <<< '%s'", u.Password)
	err := exec.Command("bash", "-c", initSudoCmd).Run()

	if err != nil {
		return err
	}

	return nil
}

// ResetSudo removes the sudo timestamp, resetting the permissions.
func (u *UserInfo) ResetSudo() error {
	err := exec.Command("bash", "-c", "sudo -K").Run()
	if err != nil {
		return err
	}

	return nil
}
