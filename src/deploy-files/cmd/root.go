package cmd

import (
	"fmt"
	"macos-deployment/deploy-files/core"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	requests "macos-deployment/deploy-files/server-requests"
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
	filevaultKey    string
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

		// initializations for structs
		filevault := core.NewFileVault(&root.config.Admin, root.script, root.log)
		user := core.NewUser(root.config.Admin, root.log)

		root.startAccountCreation(user, filevault, root.AdminStatus)
		err = root.log.WriteFile()
		if err != nil {

		}

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

		// payload for sending to the server
		logPayload := requests.NewLogPayload(root.log.GetLogName())
		// initialized with an empty key, Key gets updated below.
		// used for sending the log over to the server.
		filevaultPayload := requests.NewFileVaultPayload("")

		if root.config.FileVault {
			fvKey := root.startFileVault(filevault)

			filevaultPayload.Key = fvKey
			filevaultPayload.SetBody(root.metadata.SerialTag)
		}

		if root.config.Firewall {
			firewall := core.NewFirewall(root.log)

			root.startFirewall(firewall)
		}

		root.log.Info.Println("Sending log file to the server")

		err = root.log.WriteFile()
		if err != nil {

		}

		logPayload.Body = string(root.log.GetContent())
		err = root.startRequest(logPayload, "/api/log")
		if err != nil {
			root.log.Error.Println(fmt.Sprintf("Failed to send to data to server: %v", err))

			if filevaultPayload.Key != "" {
				root.log.Error.Println("The log file must be saved in order not to lose the generated FileVault key")
			}

			return
		}

		if filevaultPayload.Key != "" {
			root.log.Info.Println("Sending FileVault key to the server")

			err = root.startRequest(filevaultPayload, "/api/fv")
			if err != nil {
				root.log.Error.Println(fmt.Sprintf("Failed to send to data to server: %v", err))
				root.log.Error.Println("The log file must be saved in order not to lose the generated FileVault key")

				return
			}
		}

		if root.config.AlwaysCleanup {

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
func (r *RootData) startAccountCreation(user *core.UserMaker, filevault *core.FileVault, adminStatus bool) {
	if len(r.config.Accounts) < 1 {
		r.log.Warn.Println("No account information given in YAML file")
	}

	for key := range r.config.Accounts {
		currAccount := r.config.Accounts[key]

		internalUsername, err := user.CreateAccount(currAccount, adminStatus)
		if err != nil {
			// if user creation is skipped then dont log the error
			if !strings.Contains(err.Error(), "skipped") {
				logMsg := fmt.Sprintf("Error making user: %v", err)
				root.log.Error.Println(logMsg)
			}

			continue
		}

		err = filevault.AddSecureToken(internalUsername, currAccount.Password)
		if err != nil {
			r.log.Error.Println(fmt.Sprintf("Failed to add user to secure token, manual interaction needed"))

			// REMOVE THE USER, this will cause a major issue if the user does not have secure token enabled.
			// i found this out the hard way in a prod environment...
			err = user.DeleteAccount(internalUsername)
			if err != nil {
				r.log.Error.Println(fmt.Sprintf("Failed to run user removal command, manual deletion needed: %v", err))
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

// startFileVault begins the FileVault process and returns the generated key.
func (r *RootData) startFileVault(filevault *core.FileVault) string {
	fvKey := ""

	// doesn't matter if it fails, an attempt will always occur.
	fvStatus, err := filevault.Status()
	if err != nil {
		r.log.Warn.Println(fmt.Sprintf("Failed to check FileVault status: %v", err))
	}

	if !fvStatus {
		fvKey = filevault.Enable(r.config.Admin.Username, r.config.Admin.Password)
	}

	return fvKey
}

func (r *RootData) startFirewall(firewall *core.Firewall) {
	fwStatus, err := firewall.Status()
	if err != nil {
		firewallErrMsg := strings.TrimSpace(fmt.Sprintf("Failed to execute Firewall script | %v", err))
		root.log.Error.Println(firewallErrMsg, 3)
	}

	root.log.Debug.Println(fmt.Sprintf("Firewall status: %t", fwStatus))

	if !fwStatus && err == nil {
		err = firewall.Enable()
		if err != nil {
			root.log.Error.Println(err)
		}
	}
}

// startRequest sends the logs to the server.
func (r *RootData) startRequest(payload requests.Payload, endpoint string) error {
	host := r.config.ServerHost + endpoint

	serverStatus, err := requests.VerifyConnection(host)
	if err != nil {
		r.log.Error.Println(fmt.Sprintf("Error reaching host: %v", err))

		return err
	}

	if serverStatus {
		res, err := requests.POSTData(host, payload)
		if err != nil {
			return err
		}

		if !strings.Contains(res.Status, "success") {
			r.log.Warn.Println("Failed to send payload to server")
		}

	} else {
		r.log.Warn.Println("Unable to connect to host, manual interactions needed")
	}

	return nil
}
