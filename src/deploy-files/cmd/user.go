package cmd

import (
	"errors"
	"fmt"
	"os"

	embedhandler "github.com/bobllor/macdeploy/src/config"
	"github.com/bobllor/macdeploy/src/deploy-files/core"
	"github.com/bobllor/macdeploy/src/deploy-files/logger"
	"github.com/bobllor/macdeploy/src/deploy-files/scripts"
	"github.com/bobllor/macdeploy/src/deploy-files/utils"
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
	Use:   "user [command]",
	Short: "Local user commands",
}

func InitializeUserCmd() {
	initializeUserCreateCmd()

	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userAdminGrantCmd)
}

var userCreateCmd = &cobra.Command{
	Use:   "create [flags]",
	Short: "Create a local user",
	Long: "Creates a new local user on the device." +
		"\nThe created account will automatically be added to the list of " +
		"authorized users for FileVault.",
	Run: func(cmd *cobra.Command, args []string) {
		adminInfo, err := newAdminInfo()
		if err != nil {
			fmt.Printf("Failed to set admin info: %v\n", err)
			os.Exit(1)
		}

		usermaker, fv, log, file := newUserCmdStructs(adminInfo)
		// TODO: log level flag
		if file != nil {
			defer file.Close()
		}

		config, err := yaml.NewConfig(embedhandler.YAMLBytes)
		if err != nil {
			log.Warnf("Failed to read config: %v", err)
		}

		username, err := usermaker.CreateAccount(&userCobra.UserInfo, userCobra.Admin)
		if err != nil {
			log.Criticalf("Failed to create user due to an error: %v", err)
			fmt.Printf("Failed to create user %s\n", username)
			os.Exit(1)
		}

		err = fv.AddSecureToken(username, userCobra.UserInfo.Password)
		if err != nil {
			fmt.Println("Failed to add secure token to user")
			log.Warnf("Failed to add secure token for user %s: %v", username, err)
			err := usermaker.DeleteAccount(username)
			if err != nil {
				fmt.Printf("Failed to delete user %s\n", username)
				log.Warnf("Failed to delete user: %v", err)
			}

			os.Exit(1)
		}

		if userCobra.UserInfo.ApplyPolicy && config.Policy.ChangeOnLogin {
			err := usermaker.AddPasswordPolicy(username)
			if err != nil {
				log.Warnf("Failed to add password policy to %s: %v", username, err)
				os.Exit(1)
			}

			fmt.Printf("Added password policy for %s\n")
		} else if config.Policy.ChangeOnLogin {
			log.Warnf("Key 'change_on_login' is %v, unable to apply policy", config.Policy.ChangeOnLogin)
			fmt.Println("Key [policies] in YAML config requires 'change_on_login' to be true")
		}
	},
}

func initializeUserCreateCmd() {
	userCreateCmd.Flags().BoolVarP(&userCobra.Admin, "admin", "a", false, "Grants admin to the user")

	// IgnoreAdmin is not used here since user creation is manual
	userCreateCmd.Flags().StringVarP(&userCobra.UserInfo.Username, "username", "u", "", "The username of the user")
	// not recommended due to it being plain text, this is best left empty. but its an option!
	userCreateCmd.Flags().StringVarP(&userCobra.UserInfo.Password, "password", "p", "", "The password of the user")
	userCreateCmd.Flags().BoolVar(
		&userCobra.UserInfo.ApplyPolicy,
		"applypolicy",
		false,
		"Applies a password policy on login, requires options in config defined",
	)
}

var userAdminGrantCmd = &cobra.Command{
	Use:   "grantadmin <user> <user...>",
	Short: "Grants admin privileges to the given user",
	Long:  "Grants admin privileges to the given user or users",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Missing user operand\nTry 'macdeploy user admin grant -h' for more information")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		adminInfo, err := newAdminInfo()
		if err != nil {
			fmt.Printf("Failed to set admin info: %v\n", err)
			os.Exit(1)
		}

		um, _, log, file := newUserCmdStructs(adminInfo)
		if file != nil {
			defer file.Close()
		}

		for _, arg := range args {
			err := um.GrantAdmin(arg)
			if err != nil {
				log.Warnf("User argument %s got an error while granting admin: %v", arg, err)
				fmt.Printf("Failed to grant admin to user %s: %v\n", arg, err)
			} else {
				fmt.Printf("Granted admin to user %s", arg)
				log.Infof("User %s granted admin", arg)
			}
		}
	},
}

// newUserCmdStructs creates and returns four structs:
//  1. UserMaker: initialized for user related tasks
//  2. FileVault: initialized for FileVault related tasks
//  3. Logger: initialized to the default logging location or Stdout (if error)
//  4. File: The log file, this can be nil if an error occurs which must be handled
func newUserCmdStructs(adminInfo *yaml.UserInfo) (*core.UserMaker, *core.FileVault, *logger.Logger, *os.File) {
	logDir := fmt.Sprintf("%s/%s", utils.GetCurrOrHomePath(), defaultLogDir)
	// TODO: log level flag
	log, file, err := logger.NewLoggerFile(logDir, "macdeploy.user", logger.Lsilent)
	if err != nil {
		log = logger.NewStdoutLogger(logger.Lsilent)
	}

	bashScripts := scripts.NewScript()
	um := core.NewUser(*adminInfo, bashScripts, log)
	fv := core.NewFileVault(*adminInfo, bashScripts, log)

	return um, fv, log, file
}

// newAdminInfo creates a new UserInfo for admin information.
//
// Once the values are given, it will attempt to initialize sudo. If it fails,
// an error will return.
// Other errors can occur during username and password setup.
func newAdminInfo() (*yaml.UserInfo, error) {
	adminInfo := &yaml.UserInfo{}
	err := adminInfo.SetUsername()
	if err != nil {
		err := adminInfo.SetUsernameManual()
		if err != nil {
			return nil, errors.New("failed to set admin username")
		}
	}

	fmt.Println("Admin password required for sudo")
	err = adminInfo.SetPassword(false)
	if err != nil {
		return nil, errors.New("failed to set admin password")
	}
	err = adminInfo.InitializeSudo()
	if err != nil {
		return nil, errors.New("failed to initialize sudo")
	}

	return adminInfo, nil
}
