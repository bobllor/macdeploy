package cmd

import (
	"fmt"
	"macos-deployment/deploy-files/core"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"macos-deployment/deploy-files/utils"
	"macos-deployment/deploy-files/yaml"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type RootData struct {
	AdminStatus     bool
	ExcludePackages []string
	IncludePackages []string
	log             *logger.Log
	config          *yaml.Config
	script          *scripts.BashScripts
	metadata        *utils.Metadata
}

var root RootData

var rootCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(root.AdminStatus, root.ExcludePackages, root.IncludePackages)
		root.log.Info.Println(fmt.Sprintf("Starting deployment for %s", root.metadata.SerialTag))

		// initializes sudo for automation purposes.
		err := utils.InitializeSudo(root.config.Admin.Password)
		if err != nil {
			root.log.Warn.Println("Failed to authenticate sudo: %v", err)
		}

		user := core.NewUser(root.config.Admin, root.log)
		root.startAccountCreation(user, root.AdminStatus)

		// creating the files found in the search directories, it is flattened.
		searchingFiles := make([]string, 0)
		for _, searchDir := range root.config.SearchDirectories {
			searchFiles, err := utils.GetSearchFiles(searchDir)
			if err != nil {
				msg := fmt.Sprintf("Path %s does not exist, skipping path", searchDir)
				root.log.Warn.Println(msg)
				continue
			}

			searchingFiles = append(searchingFiles, searchFiles...)
		}

		root.log.Debug.Println(
			fmt.Sprintf("File amount: %d | Directories: %v", len(searchingFiles), root.config.SearchDirectories))

		if len(searchingFiles) > 0 {
			packager := core.NewPackager(root.config.Packages, searchingFiles, root.log)
			root.startPackageInstallation(packager)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// InitializeRoot initializes the flags and instances for the use of the tool.
func InitializeRoot(
	logger *logger.Log, yamlConfig *yaml.Config,
	scripts *scripts.BashScripts, metadata *utils.Metadata) {
	rootCmd.Flags().BoolVarP(
		&root.AdminStatus, "admin", "a", false, "Grants admin to the created user")
	rootCmd.Flags().StringArrayVar(&root.ExcludePackages,
		"exclude", []string{}, "Exclude a package from installation if it is defined in the YAML")
	rootCmd.Flags().StringArrayVar(&root.IncludePackages,
		"include", []string{}, "Include a package to install")

	root.log = logger
	root.config = yamlConfig
	root.script = scripts
	root.metadata = metadata
}

// accountCreation starts the account making process.
//
// User is used to call and start the account creation process.
func (r *RootData) startAccountCreation(user *core.UserMaker, adminStatus bool) {
	if len(r.config.Accounts) < 1 {
		r.log.Warn.Println("No account information given in YAML file")
	}

	for key := range r.config.Accounts {
		currAccount := r.config.Accounts[key]

		err := user.CreateAccount(currAccount, adminStatus)
		if err != nil {
			// if user creation is skipped then dont log the error
			if !strings.Contains(err.Error(), "skipped") {
				logMsg := fmt.Sprintf("Error making user: %v", err)
				root.log.Error.Println(logMsg)
			}
		}
	}
}

// startPackageInstallation begins the package installation process.
func (r *RootData) startPackageInstallation(packager *core.Packager) {
	err := packager.InstallRosetta()
	if err != nil {
		r.log.Error.Println(fmt.Sprintf("Failed to install Rosetta: %v", err))
		return
	}

	packager.RemovePackages(root.ExcludePackages)
	packager.AddPackages(root.IncludePackages)

	packages, err := packager.GetPackages(r.metadata.DistDirectory, r.script.FindPackages)
	if err != nil {
		r.log.Error.Println(fmt.Sprintf("Issue occurred with searching directory %s: %v", r.metadata.DistDirectory, err))
		return
	}

	if len(packages) < 2 {
		r.log.Warn.Println("No packages found")
		return
	}

	packager.InstallPackages(packages)
}
