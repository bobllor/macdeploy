package core

import (
	"bufio"
	"fmt"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"macos-deployment/deploy-files/utils"
	"os"
	"os/exec"
	"strconv"
)

// CreateAccount creates the user account on the device.
// Returns true if the account is successfully made, else false.
func CreateAccount(user utils.User, isAdmin bool) bool {
	// userName will be used for both entries needed.
	userName := user.User_Name

	if userName == "" {
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

		userName = utils.FormatName(input)
	} else {
		userName = utils.FormatName(userName)
	}

	initLog := fmt.Sprintf("Creating user %s", userName)
	logger.Log(initLog, 6)

	admin := "false"
	if isAdmin && !user.Ignore_Admin {
		admin = strconv.FormatBool(isAdmin)
	}

	// CreateUserScript takes 3 arguments.
	out, err := exec.Command("sudo", "bash", "-c", scripts.CreateUserScript, userName, user.Password, admin).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		errMsg := fmt.Sprintf("Failed to create user %s | Script exit status: %v", userName, err)
		logger.Log(errMsg, 3)
		return false
	}

	createdLog := fmt.Sprintf("User %s created", userName)
	logger.Log(createdLog, 6)

	return true
}
