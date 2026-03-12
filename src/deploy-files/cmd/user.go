package cmd

import (
	"fmt"
	"os"

	"github.com/bobllor/macdeploy/src/deploy-files/logger"
	"github.com/bobllor/macdeploy/src/deploy-files/yaml"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(userCmd)
}

type UserData struct {
	Admin    bool
	UserInfo yaml.UserInfo
	logvars  LogVars
}

var userCobra UserData

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Create a local user",
	PreRun: func(cmd *cobra.Command, args []string) {
		root.initialize(true)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if root.osFile != nil {
			defer root.osFile.Close()
		}

		// have to use the root from root.go, there is an
		// invalid memory address if using a new RootData.
		if userCobra.logvars.Verbose {
			root.log.SetLogLevel(logger.Linfo)
		} else if userCobra.logvars.Debug {
			root.log.SetLogLevel(logger.Ldebug)
		}

		// sets the Password if it is empty
		username, err := root.dep.usermaker.CreateAccount(&userCobra.UserInfo, userCobra.Admin)
		if err != nil {
			root.log.Criticalf("Failed to create user due to an error: %v", err)
			fmt.Printf("Failed to create user %s\n", username)
			os.Exit(1)
		}

		// will try to add to filevault regardless.
		root.postAccountCreation(username, userCobra.UserInfo.Password, userCobra.UserInfo.ApplyPolicy)
	},
}

func InitializeUserCmd() {
	// IgnoreAdmin is not used here since user creation is manual
	userCmd.Flags().StringVarP(&userCobra.UserInfo.Username, "username", "u", "", "The username of the user")
	// not recommended due to it being plain text, this is best left empty. but its an option!
	userCmd.Flags().StringVarP(&userCobra.UserInfo.Password, "password", "p", "", "The password of the user")
	userCmd.Flags().BoolVar(
		&userCobra.UserInfo.ApplyPolicy,
		"applypolicy",
		false,
		"Applies a password policy on login, requires options in config defined",
	)

	userCmd.Flags().BoolVarP(&userCobra.Admin, "admin", "a", false, "Grants admin to the user")
	userCmd.Flags().BoolVarP(&userCobra.logvars.Verbose, "verbose", "v", false, "Enables info logging")
	userCmd.Flags().BoolVar(&userCobra.logvars.Debug, "debug", false, "Enables debug and info logging")
}
