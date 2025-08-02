package core

import (
	"bufio"
	"fmt"
	"macos-deployment/deploy_files/logger"
	"macos-deployment/deploy_files/utils"
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
		input = input[:len(input)-1]
		_, validName := utils.ValidateName(input)

		if input == "" || !validName {
			println("[INFO] ")
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

	var CreateUserScript string = utils.Home + "/macos-deployment/deploy_files/create_user.sh"

	// CreateUserScript takes 3 arguments.
	_, err := exec.Command("sudo", "bash", CreateUserScript, userName, user.Password, admin).Output()
	if err != nil {
		logger.Log(err.Error(), 3)
		return false
	}

	createdLog := fmt.Sprintf("User %s created", userName)
	logger.Log(createdLog, 6)

	return true
}
