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

// NewUser creates a new User to handle user creation.
func NewUser(adminInfo yaml.UserInfo, logger *logger.Log) *UserMaker {
	user := UserMaker{
		adminInfo: adminInfo,
		log:       logger,
	}

	return &user
}

// CreateAccount creates the user account on the device.
// It will return nil upon success, if there are any errors then an error is returned.
//
// SecureToken is enabled for every user and if enabled, the user will require a password reset upon login.
func (u *UserMaker) CreateAccount(user yaml.UserInfo, isAdmin bool) error {
	// username will be used for both entries needed.
	username := user.Username

	if user.Password == "" {
		return errors.New("empty password given for user, config file must be checked")
	}

	if username == "" {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("\nHit enter to skip the user creation.")
		fmt.Print("Enter the client's name: ")

		input, _ := reader.ReadString('\n')

		fmt.Println("") // for formatting purposes.

		input = input[:len(input)-1]

		if input == "" {
			u.log.Info.Println("User creation skipped")
			return errors.New("user creation skipped")
		}

		username = input
	}

	// follows apple's naming convention
	fullName := utils.FormatFullName(username)

	initInfoLog := fmt.Sprintf("Creating user %s | Home Directory Name %s", username, fullName)
	u.log.Info.Println(initInfoLog)

	admin := "false"
	if isAdmin && !user.IgnoreAdmin {
		admin = strconv.FormatBool(isAdmin)
	}

	userExists, err := u.userExists(username)
	if err != nil {
		return fmt.Errorf("error occurred reading user directory %s: %v", username, err)
	}
	if userExists {
		return fmt.Errorf("user %s already exists in the system", username)
	}

	// CreateUserScript takes 3 arguments.
	out, err := exec.Command("sudo", "bash", "-c",
		scripts.CreateUserScript, username, fullName, user.Password, admin).CombinedOutput()
	if err != nil {
		u.log.Debug.Println(fmt.Sprintf("create user script error: %s", string(out)))
		return fmt.Errorf("failed to create user %s: %v", username, err)
	}

	// turns out i forgot secure token access... that was rough to find out in prod
	// always make securetoken, even if filevault is not enabled.
	err = addSecureToken(username, user.Password, u.adminInfo)
	if err != nil {
		return err
	}

	createdLog := fmt.Sprintf("User %s created", username)
	u.log.Info.Println(createdLog)

	if user.ChangePassword {
		pwPolicyCmd := fmt.Sprintf("sudo pwpolicy -u '%s' -setpolicy 'newPasswordRequired=1'", fullName)

		err = exec.Command("bash", "-c", pwPolicyCmd).Run()
		if err != nil {
			return fmt.Errorf("failed to create user policy for %s: %v", username, err)
		} else {
			u.log.Info.Println(fmt.Sprintf("Added new password policy for %s", username))
		}
	}

	return nil
}

func (u *UserMaker) userExists(username string) (bool, error) {
	usersPath := "/Users"

	dirs, err := os.ReadDir(usersPath)
	if err != nil {
		u.log.Error.Println(fmt.Sprintf("Error reading directory: %v", err))
		return false, err
	}

	u.log.Debug.Println(fmt.Sprintf("User directory content: %v", dirs))

	for _, dir := range dirs {
		dirName := strings.ToLower(dir.Name())
		lowerUsername := strings.ToLower(username)

		if strings.Contains(dirName, lowerUsername) {
			return true, nil
		}
	}

	return false, nil
}
