package core

import (
	"bufio"
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
// Returns true if the account is successfully made, else false.
//
// SecureToken is enabled for every user created.
// If enabled, the user will require a password reset upon login.
func CreateAccount(user yaml.User, adminInfo yaml.User, isAdmin bool) bool {
	// username will be used for both entries needed.
	username := user.User_Name

	if user.Password == "" {
		logger.Log("Cannot have an empty password for the user.", 3)
		return false
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
			return false
		}

		username = input
	}

	fullName := utils.FormatFullName(username)

	initLog := fmt.Sprintf("Creating user %s | Home Directory Name %s", username, fullName)
	logger.Log(initLog, 6)

	admin := "false"
	if isAdmin && !user.Ignore_Admin {
		admin = strconv.FormatBool(isAdmin)
	}

	userExists, err := userExists(username)
	if err != nil {
		// this is going to assume the user exists
		errMsg := fmt.Sprintf("Failed to read user directory: %s", err.Error())
		logger.Log(errMsg, 3)
	}
	if userExists {
		logger.Log(fmt.Sprintf("User %s already exists", username), 6)
		return false
	}

	// CreateUserScript takes 3 arguments.
	out, err := exec.Command("sudo", "bash", "-c",
		scripts.CreateUserScript, username, fullName, user.Password, admin).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		errMsg := fmt.Sprintf("Failed to create user %s | Script exit status: %v", username, err)
		logger.Log(errMsg, 3)
		return false
	}

	// turns out i forgot secure token access... that was rough to find out in prod
	secureTokenCmd := fmt.Sprintf("sudo sysadminctl -secureTokenOn '%s' -password '%s' -adminUser '%s' -adminPassword '%s'",
		username, user.Password, adminInfo.User_Name, adminInfo.Password)

	_, err = exec.Command("bash", "-c", secureTokenCmd).Output()
	if err != nil {
		logger.Log(fmt.Sprintf("Error enabling token for user, manual interaction needed: %v", err), 3)
	} else {
		logger.Log(fmt.Sprintf("Secure token added for %s", username), 6)
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
			logger.Log(fmt.Sprintf("Failed password policy for %s: %v", fullName, err), 3)
		} else {
			logger.Log(fmt.Sprintf("Added new password policy for %s", username), 6)
		}
	}

	return true
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
