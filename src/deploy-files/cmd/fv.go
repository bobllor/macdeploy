package cmd

import (
	"fmt"

	requests "github.com/bobllor/macdeploy/src/deploy-files/server-requests"
	"github.com/bobllor/macdeploy/src/deploy-files/yaml"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fvCmd)
	fvCmd.AddCommand(fvEnableCmd, fvDisableCmd, fvListCmd, fvStatusCmd)
}

type FileVaultData struct {
	User    yaml.UserInfo
	logvars LogVars
}

var fvCobra FileVaultData

var fvLongDescription string = `
Perform FileVault operations. When this command is ran, it will attempt to communicate
back to the server to deliver information from the client.
`

var fvCmd = &cobra.Command{
	Use:   "filevault <command> [flags]",
	Long:  fvLongDescription,
	Short: "Peform FileVault operations",
}

func InitializeFileVaultCmd() {
	fvCmd.Flags().BoolVarP(&fvCobra.logvars.Verbose, "verbose", "v", false, "Enables info logging")
	fvCmd.Flags().BoolVar(&fvCobra.logvars.Debug, "debug", false, "Enables debug logging")

	// disable subcommand
	fvDisableCmd.Flags().StringVarP(&fvCobra.User.Username, "username", "u", "", "Admin user username")
	fvDisableCmd.Flags().StringVarP(&fvCobra.User.Password, "password", "p", "", "Admin user password")

	// enable subcommand
	fvEnableCmd.Flags().StringVarP(&fvCobra.User.Username, "username", "u", "", "Admin user username")
	fvEnableCmd.Flags().StringVarP(&fvCobra.User.Password, "password", "p", "", "Admin user password")

	fvCmd.MarkFlagsMutuallyExclusive("verbose", "debug")
}

var fvDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disables FileVault",
	PreRun: func(cmd *cobra.Command, args []string) {
		root.initialize(true)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if fvCobra.User.Username == "" {
			fmt.Println("No username given")
			err := fvCobra.User.SetUsername()
			if err != nil {
				fmt.Println(err)
				root.log.Fatal(err)
				return
			}
		}
		if fvCobra.User.Password == "" {
			fmt.Println("No admin password given")
			err := fvCobra.User.SetPassword()
			if err != nil {
				fmt.Println(err)
				root.log.Fatal(err)
				return
			}
		}

		status, err := root.dep.filevault.Disable(fvCobra.User.Username, fvCobra.User.Password)
		if err != nil {
			fmt.Println("An error occurred during an attempt to disable FileVault")
			root.log.Warn(err)
			return
		}

		if status {
			fmt.Println("Disabled FileVault")
		} else {
			fmt.Println("FileVault did not get disabled")
		}
	},
}

var fvEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enables FileVault",
	Long:  "Enables FileVault and sends the key to the server",
	PreRun: func(cmd *cobra.Command, args []string) {
		root.initialize(true)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if fvCobra.User.Username == "" {
			fmt.Println("No username given, setting username")
			// this is root config as startFileVault uses root.
			err := root.config.Admin.SetUsername()
			if err != nil {
				fmt.Println(err)
				root.log.Fatal(err)
				return
			}
		}
		if fvCobra.User.Password == "" {
			fmt.Println("No password given, input required")
			// this is root config as startFileVault uses root.
			err := root.config.Admin.SetPassword()
			if err != nil {
				fmt.Println(err)
				root.log.Fatal(err)
				return
			}
		}

		r := requests.NewRequest(root.log)
		key := root.startFileVault(root.dep.filevault, r)

		keyPayload := requests.NewFileVaultPayload(key)
		keyPayload.SetBody(root.metadata.SerialTag)

		err := root.startRequest(keyPayload, r, root.config.ServerHost, "/api/fv")
		if err != nil {
			root.log.Warn(err)
			root.warnFileVaultError(keyPayload)
		} else {
			fmt.Println("FileVault enabled")
		}
	},
}

var fvListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists the users added to FileVault",
	PreRun: func(cmd *cobra.Command, args []string) {
		root.initialize(true)
	},
	Run: func(cmd *cobra.Command, args []string) {
		li, err := root.dep.filevault.List()
		if err != nil {
			str := fmt.Sprintf("Failed to run command for FileVault list: %v", err)
			fmt.Println(str)
			root.log.Warn(str)
		} else {
			fmt.Println(li)
		}
	},
}

var fvStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Checks the status of FileVault",
	PreRun: func(cmd *cobra.Command, args []string) {
		root.initialize(true)
	},
	Run: func(cmd *cobra.Command, args []string) {
		stat, err := root.dep.filevault.Status()
		if err != nil {
			fmt.Println(err)
			root.log.Warn(err)
		} else {
			if stat {
				fmt.Println("FileVault is enabled")
			} else {
				fmt.Println("FileVault is disabled")
			}
		}
	},
}
