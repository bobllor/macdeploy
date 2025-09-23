package cmd

import (
	"errors"
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
	RemoveFiles     bool
	ExcludePackages []string
	IncludePackages []string
	log             *logger.Log
	config          *yaml.Config
	script          *scripts.BashScripts
	metadata        *utils.Metadata
}

var root RootData

var rootCmd = &cobra.Command{
	Use: "deploy-arm.bin [options]",
	Run: func(cmd *cobra.Command, args []string) {
		root.log.Info.Log("Starting deployment for %s", root.metadata.SerialTag)

		// initializes sudo for automation purposes.
		err := utils.InitializeSudo(root.config.Admin.Password)
		if err != nil {
			root.log.Warn.Log("Failed to authenticate sudo: %v", err)
		}

		// dependency initializations
		filevault := core.NewFileVault(&root.config.Admin, root.script, root.log)
		user := core.NewUser(root.config.Admin, root.log)

		root.startAccountCreation(user, filevault, root.AdminStatus)
		err = root.log.WriteFile()
		if err != nil {
			fmt.Printf("Failed to write to log file: %v\n", err)
		}

		// creating the files found in the search directories, it is flattened.
		searchingFiles := make([]string, 0)
		for _, searchDir := range root.config.SearchDirectories {
			searchFiles, err := utils.GetSearchFiles(searchDir)
			if err != nil {
				root.log.Warn.Log("Path %s does not exist, skipping path", searchDir)
				continue
			}

			searchingFiles = append(searchingFiles, searchFiles...)
		}

		root.log.Debug.Log("File amount: %d | Directories: %v", len(searchingFiles), root.config.SearchDirectories)

		if len(searchingFiles) > 0 {
			packager := core.NewPackager(root.config.Packages, searchingFiles, root.log)
			root.startPackageInstallation(packager)
		}
		err = root.log.WriteFile()
		if err != nil {
			fmt.Printf("Failed to write to log file: %v\n", err)
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
		err = root.log.WriteFile()
		if err != nil {
			fmt.Printf("Failed to write to log file: %v\n", err)
		}

		if root.config.Firewall {
			firewall := core.NewFirewall(root.log)

			root.startFirewall(firewall)
		}

		root.log.Info.Log("Sending log file to the server")

		err = root.log.WriteFile()
		if err != nil {
			fmt.Printf("Failed to write to log file: %v\n", err)
		}

		request := requests.NewRequest(root.log)

		logPayload.Body = string(root.log.GetContent())
		err = root.startRequest(logPayload, request, "/api/log")
		if err != nil {
			root.log.Error.Log("Failed to send to data to server: %v", err)

			if filevaultPayload.Key != "" {
				root.log.Error.Log("The log file must be saved in order not to lose the generated FileVault key")
			}

			return
		}

		if filevaultPayload.Key != "" {
			root.log.Info.Log("Sending FileVault key to the server")

			err = root.startRequest(filevaultPayload, request, "/api/fv")
			if err != nil {
				root.log.Error.Log("Failed to send to data to server: %v", err)
				fmt.Println("The log file must be saved in order not to lose the generated FileVault key")

				err = root.log.WriteFile()
				if err != nil {
					fmt.Printf("Failed to write to log file: %v\n", err)
				}

				return
			}
		}

		if root.RemoveFiles {
			filesToRemove := map[string]struct{}{
				root.metadata.DistDirectory: {},
				root.metadata.ZipFile:       {},
			}

			root.startCleanup(filesToRemove)
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
		"exclude", []string{}, "Exclude a package from installation")
	rootCmd.Flags().StringArrayVar(&root.IncludePackages,
		"include", []string{}, "Include a package to install")
	rootCmd.Flags().BoolVar(
		&root.RemoveFiles, "remove-files", false, "Removes the files on the device upon successful execution")

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
		r.log.Warn.Log("No account information given in YAML file")
	}

	for key := range r.config.Accounts {
		currAccount := r.config.Accounts[key]

		internalUsername, err := user.CreateAccount(currAccount, adminStatus)
		if err != nil {
			// if user creation is skipped then dont log the error
			if !strings.Contains(err.Error(), "skipped") {
				logMsg := fmt.Sprintf("Error making user: %v", err)
				root.log.Error.Log(logMsg)
			}

			continue
		}

		err = filevault.AddSecureToken(internalUsername, currAccount.Password)
		if err != nil {
			r.log.Error.Log("Failed to add user to secure token, manual interaction needed")

			// REMOVE THE USER, this will cause a major issue if the user does not have secure token enabled.
			// i found this out the hard way in a prod environment...
			err = user.DeleteAccount(internalUsername)
			if err != nil {
				r.log.Error.Log("Failed to run user removal command, manual deletion needed: %v", err)
			}
		}

	}
}

// startPackageInstallation begins the package installation process.
func (r *RootData) startPackageInstallation(packager *core.Packager) {
	err := packager.InstallRosetta()
	if err != nil {
		r.log.Error.Printf("Failed to install Rosetta: %v", err)
		return
	}

	packager.RemovePackages(root.ExcludePackages)
	packager.AddPackages(root.IncludePackages)

	packages, err := packager.GetPackages(r.metadata.DistDirectory, r.script.FindPackages)
	if err != nil {
		r.log.Error.Log("Issue occurred with searching directory %s: %v", r.metadata.DistDirectory, err)
		return
	}

	if len(packages) < 2 {
		r.log.Warn.Log("No packages found")
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
		r.log.Warn.Printf("Failed to check FileVault status: %v", err)
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
		root.log.Error.Printf(firewallErrMsg, 3)
	}

	root.log.Debug.Printf("Firewall status: %t", fwStatus)

	if !fwStatus && err == nil {
		err = firewall.Enable()
		if err != nil {
			root.log.Error.Log("%v", err)
		}
	}
}

// startRequest sends the logs to the server.
func (r *RootData) startRequest(payload requests.Payload, request *requests.Request, endpoint string) error {
	host := r.config.ServerHost + endpoint

	serverStatus, err := request.VerifyConnection(r.config.ServerHost)
	if err != nil {
		r.log.Error.Printf("Error reaching host: %v", err)

		return err
	}

	if serverStatus {
		res, err := request.POSTData(host, payload)
		if err != nil {
			return err
		}

		if !strings.Contains(res.Status, "success") {
			r.log.Warn.Log("Failed to send payload to server")

			return errors.New("payload failed to send to server")
		}

		r.log.Info.Log("Successfully sent log file to server")
	} else {
		r.log.Warn.Log("Unable to connect to host, manual interactions needed")

		return errors.New("unable to connect to host")
	}

	return nil
}

// startCleanup begins the cleanup process.
func (r *RootData) startCleanup(filesToRemove map[string]struct{}) {
	currDir, err := os.Getwd()
	// i am unsure what errors can happen here
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		return
	}

	currDir = strings.ToLower(currDir)

	// just in case this is ran in the main project directory.
	if strings.Contains(currDir, r.metadata.ProjectName) {
		fmt.Printf("project directory is forbidden, clean up aborted %v\n", currDir)
		return
	}

	files, err := os.ReadDir(currDir)
	if err != nil {
		fmt.Printf("Unable to read directory %v\n", err)
		return
	}

	utils.RemoveFiles(filesToRemove, files)
}
