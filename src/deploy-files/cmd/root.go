package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "command",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: fix!
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// InitializeRoot initializes the flags and instances for the use of the tool.
func InitializeRoot() {

}
