package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// unlike the main process, this will not use the packages from the config file.
// this does not support checking for installed files. it's the user's responsibility
// to know this beforehand.

func init() {
	rootCmd.AddCommand(installCmd)
}

type InstallData struct {
	packages []string
	dmg      bool
	logvars  LogVars
}

var installCobra InstallData

var longDescription string = `
Installs an argument of pkg files that are stored inside the 'dist' folder of the
deployment files.
`

var installCmd = &cobra.Command{
	Use:   "install <file> [<file>...]",
	Long:  longDescription,
	Short: "Installs packages from the 'dist' folder",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Missing file argument")
			os.Exit(1)
		}

		installCobra.packages = args
		root.initialize(true)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// have to use the root from root.go, there is an
		// invalid memory address if using a new RootData.
		if installCobra.logvars.Verbose {
			root.log.EnableInfoLog()
		} else if installCobra.logvars.Debug {
			root.log.EnableDebugLog()
		}

		// yes i know. i didnt want to rewrite a good chunk of my project so
		// why not just do it this way lol.
		root.dep.filehandler.RemovePackages(root.dep.filehandler.GetPackages())
		root.dep.filehandler.AddPackages(installCobra.packages)

		// due to the way i coded this, we will search the distribution folder first
		// for all .pkg files.
		// DMG mounts are an option flag.
		if installCobra.dmg {
			dmgFiles, err := root.dep.filehandler.ReadDir(root.metadata.DistDirectory, ".dmg")
			if err != nil {
				root.log.Error.Log(err.Error())
			} else {

				if len(dmgFiles) > 0 {
					volumeMounts := root.dep.filehandler.AttachDmgs(dmgFiles)

					root.dep.filehandler.AddDmgPackages(volumeMounts, root.metadata.DistDirectory)
					root.dep.filehandler.DetachDmgs(volumeMounts)
				} else {
					root.log.Warn.Log("No DMG files found in %s", root.metadata.DistDirectory)
				}
			}
		}

		data, err := root.dep.filehandler.ReadDir(root.metadata.DistDirectory, ".pkg")
		if err != nil {
			root.log.Error.Log(err.Error())
		} else {
			root.dep.filehandler.InstallPackages(data, []string{})
		}

	},
}

func InitializeInstallCmd() {
	installCmd.Flags().BoolVar(&installCobra.dmg, "mount-dmg", false, "Mounts and extracts the contents of DMG files automatically")
	installCmd.Flags().BoolVarP(&installCobra.logvars.Verbose, "verbose", "v", false, "Enables info logging")
	installCmd.Flags().BoolVar(&installCobra.logvars.Debug, "debug", false, "Enables debug and info logging")

	rootCmd.AddCommand(installCmd)
}
