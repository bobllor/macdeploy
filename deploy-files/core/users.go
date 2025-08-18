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
func CreateAccount(user yaml.User, isAdmin bool) bool {
	// username will be used for both entries needed.
	username := user.User_Name

	if username == "" {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("\nNaming format (case insensitive): FIRST LAST || FIRST.LAST || F LAST || F.LAST)")
		fmt.Println("Hit enter if you want to skip the user's entry.")
		fmt.Print("Enter the client's name: ")

		input, _ := reader.ReadString('\n')

		fmt.Println("") // for formatting purposes.

		input = input[:len(input)-1]
		validName := utils.ValidateName(input)

		if input == "" || !validName {
			logger.Log("User creation skipped", 6)
			return false
		}

		username = utils.FormatName(input)
	} else {
		username = utils.FormatName(username)
	}

	initLog := fmt.Sprintf("Creating user %s", username)
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
	out, err := exec.Command("sudo", "bash", "-c", scripts.CreateUserScript, username, user.Password, admin).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		errMsg := fmt.Sprintf("Failed to create user %s | Script exit status: %v", username, err)
		logger.Log(errMsg, 3)
		return false
	}

	createdLog := fmt.Sprintf("User %s created", username)
	logger.Log(createdLog, 6)

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
