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

// CreateAccount creates the user account on the device.
// It will return nil upon success, if there are any errors then an error is returned.
//
// SecureToken is enabled for every user and if enabled, the user will require a password reset upon login.
func CreateAccount(user yaml.User, adminInfo yaml.User, isAdmin bool) error {
	// username will be used for both entries needed.
	username := user.User_Name

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
			logger.Log("User creation skipped", 6)
			return errors.New("user creation skipped")
		}

		username = input
	}

	// IMPORTANT: this is the actual internal name of the user!
	// because mac is stupid and named it fullname for some reason!
	fullName := utils.FormatFullName(username)

	initLog := fmt.Sprintf("Creating user %s | Home Directory Name %s", username, fullName)
	logger.Log(initLog, 6)

	admin := "false"
	if isAdmin && !user.Ignore_Admin {
		admin = strconv.FormatBool(isAdmin)
	}

	userExists, err := userExists(fullName)
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
		logger.Log(fmt.Sprintf("user creation failed info: %s", string(out)), 7)
		return fmt.Errorf("failed to create user %s: %v", username, err)
	}

	// turns out i forgot secure token access... that was rough to find out in prod
	// always make securetoken, even if filevault is not enabled.
	err = addSecureToken(username, user.Password, adminInfo)
	if err != nil {
		return err
	}

	// msg used for logging purposes.
	createdUserString := ""
	if isAdmin && !user.Ignore_Admin {
		createdUserString = "with admin"
	}
	createdLog := fmt.Sprintf("User %s created %s", username, createdUserString)
	logger.Log(createdLog, 6)

	if user.Change_Password {
		pwPolicyCmd := fmt.Sprintf("sudo pwpolicy -u '%s' -setpolicy 'newPasswordRequired=1'", fullName)

		err = exec.Command("bash", "-c", pwPolicyCmd).Run()
		if err != nil {
			return fmt.Errorf("failed to create user policy for %s: %v", username, err)
		} else {
			logger.Log(fmt.Sprintf("Added new password policy for %s", username), 6)
		}
	}

	return nil
}

func userExists(username string) (bool, error) {
	usersPath := "/Users"

	dirs, err := os.ReadDir(usersPath)
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading directory: %s", err.Error()), 3)
		return false, err
	}

	logger.Log(fmt.Sprintf("User directory content: %v", dirs), 7)

	for _, dir := range dirs {
		dirName := strings.ToLower(dir.Name())
		lowerUsername := strings.ToLower(username)

		if strings.Contains(dirName, lowerUsername) {
			return true, nil
		}
	}

	return false, nil
}

// moveToDesktop moves a given file to the user's desktop.
func moveToDesktop(fileName string, user string) error {
	_, err := os.Stat(fileName)
	if err != nil {
		return err
	}

	moveCmd := fmt.Sprintf("sudo cp %s /Users/%s/desktop", fileName, user)
	err = exec.Command("bash", "-c", moveCmd).Run()
	if err != nil {
		return err
	}

	logger.Log(fmt.Sprintf("Moved %s to %s", fileName, user), 6)

	return nil
}
