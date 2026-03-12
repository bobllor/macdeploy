package cmd

import (
	"fmt"
	"os"

	"github.com/bobllor/macdeploy/src/deploy-files/logger"

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
	Use:   "install <file> [<file>...] [flags]",
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
		if root.osFile != nil {
			defer root.osFile.Close()
		}

		// have to use the root from root.go, there is an
		// invalid memory address if using a new RootData.
		if installCobra.logvars.Verbose {
			root.log.SetLogLevel(logger.Linfo)
		} else if installCobra.logvars.Debug {
			root.log.SetLogLevel(logger.Ldebug)
		}

		// yes i know. i didnt want to rewrite a good chunk of my project so
		// why not just do it this way lol.
		root.log.Info("Removing listed config packages")
		root.dep.filehandler.RemovePackages(root.dep.filehandler.GetPackages())
		root.dep.filehandler.AddPackages(installCobra.packages)

		// due to the way i coded this, we will search the distribution folder first
		// for all .pkg files.
		// DMG mounts are an option flag.
		if installCobra.dmg {
			dmgFiles, err := root.dep.filehandler.ReadDir(root.metadata.Files.DistDirectory, ".dmg")
			if err != nil {
				root.log.Warn(err.Error())
				fmt.Printf("Could not find folder '%s' for DMG files\n", root.metadata.Files.DistDirectory)
			} else {
				if len(dmgFiles) > 0 {
					volumeMounts := root.dep.filehandler.AttachDmgs(dmgFiles)

					root.dep.filehandler.AddDmgPackages(volumeMounts, root.metadata.Files.DistDirectory)
					root.dep.filehandler.DetachDmgs(volumeMounts)
				} else {
					root.log.Warn(fmt.Sprintf("No DMG files found in %s", root.metadata.Files.DistDirectory))
				}
			}
		}

		data, err := root.dep.filehandler.ReadDir(root.metadata.Files.DistDirectory, ".pkg")
		if err != nil {
			root.log.Warn(err.Error())
			fmt.Printf("Could not find folder '%s' for PKG files\n", root.metadata.Files.DistDirectory)
		} else {
			root.dep.filehandler.InstallPackages(data, []string{})
		}
	},
}

func InitializeInstallCmd() {
	installCmd.Flags().BoolVar(&installCobra.dmg, "mountdmg", false, "Mounts and extracts the contents of DMG files")
	installCmd.Flags().BoolVarP(&installCobra.logvars.Verbose, "verbose", "v", false, "Enables info logging")
	installCmd.Flags().BoolVar(&installCobra.logvars.Debug, "debug", false, "Enables debug logging")
}
