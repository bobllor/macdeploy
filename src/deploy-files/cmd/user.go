package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	embedhandler "github.com/bobllor/macdeploy/src/config"
	"github.com/bobllor/macdeploy/src/deploy-files/core"
	"github.com/bobllor/macdeploy/src/deploy-files/logger"
	"github.com/bobllor/macdeploy/src/deploy-files/scripts"
	"github.com/bobllor/macdeploy/src/deploy-files/utils"
	"github.com/bobllor/macdeploy/src/deploy-files/yaml"

	"github.com/spf13/cobra"
)

const userLogName = "macdeploy.user"

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

	userCmd.PersistentFlags().BoolVar(&userCobra.logvars.Verbose, "verbose", false, "Show info level logging")
	userCmd.PersistentFlags().BoolVar(&userCobra.logvars.Debug, "debug", false, "Show debug level logging")

	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userDeleteCmd)
	userCmd.AddCommand(userAdminGrantCmd)
	userCmd.AddCommand(userAdminRevokeCmd)
	userCmd.AddCommand(userListCmd)
}

var userCreateCmd = &cobra.Command{
	Use:   "create [<user>] [flags]",
	Short: "Create a local user",
	Long: "Creates a new local user on the device." +
		"\nThe created account will automatically be added to the list of " +
		"authorized users for FileVault.",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			userCobra.UserInfo.Username = args[0]
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		adminInfo, err := newAdminInfo()
		if err != nil {
			fmt.Println("Failed to retrieve admin information")
			os.Exit(1)
		}
		logLevel := getLogLevel(userCobra.logvars)

		usermaker, fv, log, file := newUserCmdStructs(adminInfo, logLevel)
		if file != nil {
			defer file.Close()
		}

		// used for applying password policies
		// manual password policies are not supported at the time (5/8/2026)
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
		if userCobra.Admin {
			log.Infof("Admin flag used for %s", userCobra.UserInfo.Username)
			fmt.Printf("User %s granted admin\n", userCobra.UserInfo.Username)
		}

		err = fv.AddSecureToken(username, userCobra.UserInfo.Password)
		if err != nil {
			fmt.Println("Failed to add secure token to user")
			log.Warnf("Failed to add secure token for user %s: %v", username, err)
			err := usermaker.DeleteAccount(username)
			if err != nil {
				log.Warnf("Failed to delete user: %v", err)
				fmt.Printf("Failed to delete user %s\n", username)
			}

			os.Exit(1)
		}

		if userCobra.UserInfo.ApplyPolicy && config.Policy.ChangeOnLogin {
			err := usermaker.AddPasswordPolicy(username)
			if err != nil {
				log.Warnf("Failed to add password policy to %s: %v", username, err)
				os.Exit(1)
			}

			fmt.Printf("Added password policy for %s\n", username)
		} else if userCobra.UserInfo.ApplyPolicy && !config.Policy.ChangeOnLogin {
			log.Warnf("Key 'change_on_login' is %v, unable to apply policy", config.Policy.ChangeOnLogin)
			fmt.Println("Key [policies] in YAML config requires 'change_on_login' to be true to set the policy")
		}
	},
}

func initializeUserCreateCmd() {
	userCreateCmd.Flags().BoolVarP(&userCobra.Admin, "admin", "a", false, "Grants admin to the user")

	// not recommended due to it being plain text, this is best left empty. but its an option!
	userCreateCmd.Flags().StringVarP(&userCobra.UserInfo.Password, "password", "p", "", "The password of the user")
	userCreateCmd.Flags().BoolVar(
		&userCobra.UserInfo.ApplyPolicy,
		"applypolicy",
		false,
		"Applies a password policy on login, requires options in config defined",
	)
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete <user> [<user>...]",
	Short: "Deletes a user",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Missing user operand\nTry 'macdeploy user delete -h' for more information")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		adminInfo, err := newAdminInfo()
		if err != nil {
			fmt.Println("Failed to retrieve admin information")
			os.Exit(1)
		}

		um, _, log, file := newUserCmdStructs(adminInfo, getLogLevel(userCobra.logvars))
		if file != nil {
			defer file.Close()
		}

		for _, arg := range args {
			err = um.DeleteAccount(arg)
			if err != nil {
				log.Warnf("Failed to delete user %s: %v", arg, err)
				fmt.Printf("Failed to delete user %s\n", arg)
				continue
			}

			fmt.Printf("Deleted user %s\n", arg)
			log.Infof("Deleted user %s", arg)
		}
	},
}

var userAdminGrantCmd = &cobra.Command{
	Use:   "grantadmin <user> [<user>...]",
	Short: "Grants admin privileges to users",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Missing user operand\nTry 'macdeploy user grantadmin -h' for more information")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		adminInfo, err := newAdminInfo()
		if err != nil {
			fmt.Println("Failed to retrieve admin information")
			os.Exit(1)
		}
		logLevel := getLogLevel(userCobra.logvars)

		um, _, log, file := newUserCmdStructs(adminInfo, logLevel)
		if file != nil {
			defer file.Close()
		}

		for _, arg := range args {
			user := utils.FormatUsername(arg)
			err := um.GrantAdmin(user)
			if err != nil {
				log.Warnf("User argument %s (%s) got an error while granting admin: %v", arg, user, err)
				fmt.Printf("Failed to grant admin to user %s\n", arg)
				continue
			}

			fmt.Printf("Granted admin to user %s\n", arg)
			log.Infof("User %s granted admin", arg)
		}
	},
}

var userAdminRevokeCmd = &cobra.Command{
	Use:   "revokeadmin <user> [<user>...]",
	Short: "Revokes admin privileges to the given user",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Missing user operand\nTry 'macdeploy user revokeadmin -h' for more information")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		adminInfo, err := newAdminInfo()
		if err != nil {
			fmt.Println("Failed to retrieve admin information")
			os.Exit(1)
		}
		logLevel := getLogLevel(userCobra.logvars)

		um, _, log, file := newUserCmdStructs(adminInfo, logLevel)
		if file != nil {
			defer file.Close()
		}

		for _, arg := range args {
			user := utils.FormatUsername(arg)
			err := um.RevokeAdmin(user)
			if err != nil {
				log.Warnf("User argument %s got an error while revoking admin: %v", arg, err)
				fmt.Printf("Failed to revoke admin from user %s\n", arg)
			} else {
				fmt.Printf("Revoked admin to user %s\n", arg)
				log.Infof("User %s revoked admin", arg)
			}
		}
	},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists the local users on the device",
	Run: func(cmd *cobra.Command, args []string) {
		logLevel := getLogLevel(userCobra.logvars)
		log, file, err := logger.NewLoggerFile(utils.GetCurrOrHomePath()+"/"+defaultLogDir, userLogName, logLevel)
		if err != nil {
			log = logger.NewStdoutLogger(logLevel)
		}
		if file != nil {
			defer file.Close()
		}

		um := core.NewUser(userCobra.UserInfo, scripts.NewScript(), log)

		users, err := um.List()
		if err != nil {
			log.Warnf("Failed to get local users list: %v", err)
			fmt.Println("Failed to retrieve local users list")
		}

		fmt.Println(strings.Join(users, "\n"))
	},
}

// newUserCmdStructs creates and returns four structs:
//  1. UserMaker: initialized for user related tasks
//  2. FileVault: initialized for FileVault related tasks
//  3. Logger: initialized to the default logging location or Stdout (if error)
//  4. File: The log file, this can be nil if an error occurs which must be handled
func newUserCmdStructs(adminInfo *yaml.UserInfo, logLevel int) (*core.UserMaker, *core.FileVault, *logger.Logger, *os.File) {
	logDir := fmt.Sprintf("%s/%s", utils.GetCurrOrHomePath(), defaultLogDir)
	log, file, err := logger.NewLoggerFile(logDir, userLogName, logLevel)
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

	fmt.Println("Admin password required")
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

// getLogLevel retrieves the log level based on the flag value.
func getLogLevel(v LogVars) int {
	if v.Verbose {
		return logger.Linfo
	}
	if v.Debug {
		return logger.Ldebug
	}

	return logger.Lsilent
}
