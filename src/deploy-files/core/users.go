package core

import (
	"bufio"
	"errors"
	"fmt"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"macos-deployment/deploy-files/utils"
	"macos-deployment/deploy-files/yaml"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type UserMaker struct {
	adminInfo yaml.UserInfo
	log       *logger.Log
}

// NewUser creates a new UserMaker to handle user creation.
func NewUser(adminInfo yaml.UserInfo, logger *logger.Log) *UserMaker {
	user := UserMaker{
		adminInfo: adminInfo,
		log:       logger,
	}

	return &user
}

// CreateAccount creates the user account on the device.
//
// It will return the internal username of the macOS account upon success.
// If there are any errors then an error is returned.
func (u *UserMaker) CreateAccount(user yaml.UserInfo, isAdmin bool) (string, error) {
	// username will be used for both entries needed.
	username := user.Username

	if user.Password == "" {
		return "", errors.New("empty password given for user, config file must be checked")
	}

	if username == "" {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("\nHit enter to skip the user creation.")
		fmt.Print("Enter the client's name: ")

		input, _ := reader.ReadString('\n')

		fmt.Println("") // for formatting purposes.

		input = input[:len(input)-1]

		if input == "" {
			u.log.Info.Log("User creation skipped")
			return "", errors.New("user creation skipped")
		}

		username = input
	}

	// follows apple's naming convention
	accountName := utils.FormatFullName(username)

	initInfoLog := fmt.Sprintf("Creating user %s | Account Name %s", username, accountName)
	u.log.Info.Log(initInfoLog)

	admin := "false"
	if isAdmin && !user.IgnoreAdmin {
		u.log.Info.Log("Admin enabled for user %s", username)
		admin = strconv.FormatBool(isAdmin)
	}

	userExists, err := u.userExists(accountName)
	if err != nil {
		return "", fmt.Errorf("error occurred reading user directory %s: %v", username, err)
	}
	if userExists {
		return "", fmt.Errorf("user %s already exists in the system", username)
	}

	// CreateUserScript takes 3 arguments.
	out, err := exec.Command("sudo", "bash", "-c",
		scripts.CreateUserScript, username, accountName, user.Password, admin).CombinedOutput()
	if err != nil {
		u.log.Debug.Log(fmt.Sprintf("create user script error: %s", string(out)))
		return "", fmt.Errorf("failed to create user %s: %v", username, err)
	}

	createdLog := fmt.Sprintf("User %s created", username)
	u.log.Info.Log(createdLog)

	return accountName, nil
}

// DeleteAccount removes the given user from the device.
func (u *UserMaker) DeleteAccount(username string) error {
	cmd := fmt.Sprintf(
		"sudo sysadminctl -deleteUser '%s' -adminUser '%s' -adminPassword '%s'",
		username,
		u.adminInfo.Username,
		u.adminInfo.Password)
	_, err := exec.Command("sudo", "bash", "-c", cmd).Output()
	if err != nil {
		return err
	}

	u.log.Info.Log(fmt.Sprintf("Removed user %s", username))

	return nil
}

// AddPasswordPolicy adds a password policy for the user. This can only be
// ran after adding the Secure Token to the user.
//
// If the password policy fails to run then an error returns.
func (u *UserMaker) AddPasswordPolicy(username string) error {
	pwPolicyCmd := fmt.Sprintf("sudo pwpolicy -u '%s' -setpolicy 'newPasswordRequired=1'", username)

	err := exec.Command("bash", "-c", pwPolicyCmd).Run()
	if err != nil {
		return fmt.Errorf("failed to create user policy for %s: %v", username, err)
	}

	u.log.Info.Log(fmt.Sprintf("Added new password policy for %s", username))

	return nil
}

// userExists checks the Users directory for the given username.
func (u *UserMaker) userExists(username string) (bool, error) {
	usersPath := "/Users"

	dirs, err := os.ReadDir(usersPath)
	if err != nil {
		u.log.Error.Log(fmt.Sprintf("Error reading directory: %v", err))
		return false, err
	}

	u.log.Debug.Log(fmt.Sprintf("User directory content: %v", dirs))

	for _, dir := range dirs {
		dirName := strings.ToLower(dir.Name())
		lowerUsername := strings.ToLower(username)

		if strings.Contains(dirName, lowerUsername) {
			return true, nil
		}
	}

	return false, nil
}
