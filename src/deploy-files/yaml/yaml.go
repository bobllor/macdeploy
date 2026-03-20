package yaml

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
	"golang.org/x/term"
)

type Config struct {
	// Accounts is a map of UserInfo used to create local accounts on the device.
	Accounts map[string]UserInfo `yaml:"accounts"`

	// Packages are the package file names that are to be installed, with
	// a slice of the install file names used to conditionally
	// install packages if found in an install directory.
	Packages map[string][]string `yaml:"packages"`

	// InstallDirectories is a slice of paths that will contain the install files
	// of packages.
	InstallDirectories []string `yaml:"install_directories"`

	// Scripts is a type used to inject script execution during deployment.
	Scripts ScriptTypes `yaml:"scripts"`

	// Admin is a UserInfo type that is used for admin elevation. For security purposes
	// this can be omitted.
	Admin UserInfo

	// Policy is a type that holds the policy information to apply password policies to
	// the created user.
	Policy Policies `yaml:"policies"`

	// ServerHost is the host of the server for the deployment process. This is required and
	// must be a URL (https/http).
	ServerHost string `yaml:"server_host" validate:"url,required"`

	// FileVault is used to enable or ignore enabling FileVault.
	FileVault bool

	// Firewall is used to enable or ignore enabling Firewall.
	Firewall bool

	// Cleanup is used for confirmation before file removal. By default the value is "warn".
	// The allowed values are ["warn","force"].
	// This only effects the user prompt before a cleanup occurs, it still requires
	// the --cleanup flag to be used.
	//
	// If the value is in the allowed values then it will fail to validate.
	//
	//	- warn: Requires a confirmation before removing the deployment files.
	//	The default option.
	//	- force: Remove deployment files with no confirmation.
	//
	// The option "force" will not override the confirmation if either
	// the FileVault process or the POST to the server with the FileVault key failed.
	Cleanup string `yaml:"cleanup" validate:"oneof=warn force"`
}

type UserInfo struct {
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	IgnoreAdmin bool   `yaml:"ignore_admin"`
	ApplyPolicy bool   `yaml:"apply_policy"`
}

// ScriptTypes contains fields with string slices representing the
// script file names used to execute during the lifecycle of the process.
type ScriptTypes struct {
	Pre  []string `yaml:"pre"`
	Mid  []string `yaml:"mid"`
	Post []string `yaml:"post"`
}

// NewConfig returns a struct containing data read from the YAML file. The file is read
// through embedding.
//
// If an issue occurs while reading the file then it will return an error.
func NewConfig(data []byte) (*Config, error) {
	config := Config{
		Cleanup: "warn",
	}

	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate validates the Config structure. It will return an error
// with all the failed keys of Config for any failed validation.
func Validate(config *Config) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	configKeys := []string{
		"Cleanup",
		"ServerHost",
	}

	yamlErrHandler := NewConfigError(configKeys)

	yamlErrHandler.SetKeyError("Cleanup", "field 'cleanup' (%s) is invalid, validation failed on %s (allowed values [%s])")
	yamlErrHandler.SetKeyError("ServerHost", "field 'server_host' (%s) is invalid, validation failed on %s (https/http)")

	err := validate.Struct(config)
	if err != nil {

		errs := err.(validator.ValidationErrors)

		errBuilder := []string{}

		for _, e := range errs {
			errStr, configErr := yamlErrHandler.GetKeyError(e.Field())
			if configErr != nil {
				return configErr
			}

			param := e.Param()
			tag := e.Tag()
			val := e.Value()

			// handles Param if included
			outErr := fmt.Sprintf(errStr, val, tag)
			if param != "" {
				outErr = fmt.Sprintf(errStr, val, tag, param)
			}

			errBuilder = append(errBuilder, outErr)
		}

		return errors.New(strings.Join(errBuilder, "\n"))
	}

	return nil
}

// Marshal serializes an interface into a bytes value.
func Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

// Unmarshal serializes a bytes value into an interface.
func Unmarshal(data []byte) (any, error) {
	var v any

	err := yaml.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}

	return v, nil
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
	fmt.Print("Enter password: ")
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
		fmt.Print("Confirm password: ")
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
