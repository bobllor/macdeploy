package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(userCmd)
}

type UserData struct {
	Admin       bool
	ApplyPolicy bool
	logvars     LogVars
}

var userCobra UserData

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Create a user",
	Run: func(cmd *cobra.Command, args []string) {
		// have to use the root from root.go, there is an
		// invalid memory address if using a new RootData.
		root.initialize(true)
		if userCobra.logvars.Verbose {
			root.log.EnableInfoLog()
		} else if userCobra.logvars.Debug {
			root.log.EnableDebugLog()
		}

		root.CreateLocal = true

		root.startAccountCreation(userCobra.Admin)
	},
}

func InitializeUserCmd() {
	userCmd.Flags().BoolVarP(&userCobra.Admin, "admin", "a", false, "Grants admin to the user")
	userCmd.Flags().BoolVar(&userCobra.ApplyPolicy, "apply-policy", false, "Applies a password reset policy on login")
	userCmd.Flags().BoolVarP(&userCobra.logvars.Verbose, "verbose", "v", false, "Enables info logging")
	userCmd.Flags().BoolVar(&userCobra.logvars.Debug, "debug", false, "Enables debug and info logging")

	rootCmd.AddCommand(userCmd)
}
