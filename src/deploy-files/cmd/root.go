package cmd

import (
	"bufio"
	"errors"
	"fmt"
	embedhandler "macos-deployment/config"
	"macos-deployment/deploy-files/core"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	requests "macos-deployment/deploy-files/server-requests"
	"macos-deployment/deploy-files/utils"
	"macos-deployment/deploy-files/yaml"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

type RootData struct {
	AdminStatus     bool
	RemoveFiles     bool
	Verbose         bool
	Debug           bool
	NoSend          bool
	Mount           bool
	ExcludePackages []string
	IncludePackages []string
	PlistPath       string
	log             *logger.Log
	config          *yaml.Config
	script          *scripts.BashScripts
	metadata        *utils.Metadata
	errors          errorFlags
	data            varData
	dep             dependencies
	perm            *utils.Perms
}

type errorFlags struct {
	ServerFailed  bool // Indicates if sending the files to the server has failed.
	ScriptsFailed bool // Indicates if searching for script files (.sh) has failed.
}

type dependencies struct {
	filehandler *core.FileHandler
	usermaker   *core.UserMaker
	filevault   *core.FileVault
	firewall    *core.Firewall
}

type varData struct {
	scriptFiles []string
}

var root RootData

var rootCmd = &cobra.Command{
	Use:   "macdeploy",
	Short: "MacBook deployment tool",
	Long:  `Automated deployment for MacBooks for ITAMs.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		const projectName string = "macos-deployment"
		const distDirectory string = "dist"
		const zipFile string = "deploy.zip"

		// not exiting, just in case mac fails somehow. but there are checks for non-mac devices.
		serialTag, err := utils.GetSerialTag()
		if err != nil {
			serialTag = "UNKNOWN"
			fmt.Printf("Unable to get serial number: %v\n", err)
		}

		perms := utils.NewPerms()
		root.perm = perms

		metadata := utils.NewMetadata(projectName, serialTag, distDirectory, zipFile)
		scripts := scripts.NewScript()
		config, err := yaml.NewConfig(embedhandler.YAMLBytes)
		if err != nil {
			// TODO: make this a better error message (incorrect keys, required keys missing, etc)
			fmt.Printf("Error parsing YAML configuration, %v\n", err)
			return
		}

		// checking if admin info was given or not
		if config.Admin.Username == "" {
			err = config.Admin.SetUsername()
			if err != nil {
				fmt.Printf("Failed to get username of admin: %v\n", err)
				os.Exit(1)
			}
		}
		if config.Admin.Password == "" {
			fmt.Println("No admin password given")
			err = config.Admin.SetPassword()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		err = config.Admin.InitializeSudo()
		if err != nil {
			fmt.Printf("Failed to initialize sudo with given password: %v\n", err)
		}

		// by default we will put in the home directory if none is given
		logDirectory := config.Log
		defaultLogDir := fmt.Sprintf("%s/%s", metadata.Home, "logs/macdeploy")
		if logDirectory == "" {
			logDirectory = defaultLogDir
		} else {
			logDirectory = logger.FormatLogPath(logDirectory)
		}

		// mkdir needs full permission for some reason.
		// anything other than full will have permissions of 000.
		// full perms assigns it the normal permissions: rwxr-xr-x. which is odd to me.
		err = logger.MkdirAll(logDirectory, root.perm.Full)
		if err != nil {
			fmt.Printf("Unable to make logging directory: %v\n", err)
			fmt.Printf("Changing log output to home directory: %s\n", defaultLogDir)
			logDirectory = defaultLogDir

			err = logger.MkdirAll(defaultLogDir, root.perm.Full)
			if err != nil {
				fmt.Printf("Unable to make logging directory: %v\n", err)
			}
		}

		log := logger.NewLog(serialTag, logDirectory, root.Verbose, root.Debug)
		log.Info.Log("Starting deployment for %s", metadata.SerialTag)
		fmt.Printf("Starting deployment for %s\n", metadata.SerialTag)
		log.Debug.Log("Log directory: %s", logDirectory)

		// dependency initializations
		filevault := core.NewFileVault(config.Admin, scripts, log)
		user := core.NewUser(config.Admin, scripts, log)
		handler := core.NewFileHandler(config.Packages, log)
		firewall := core.NewFirewall(log, scripts)

		root.log = log
		root.config = config
		root.script = scripts
		root.metadata = metadata

		root.dep.usermaker = user
		root.dep.filehandler = handler
		root.dep.firewall = firewall
		root.dep.filevault = filevault

		// initialized for the lifecycle during pre, install, and post script stages
		scriptFiles, err := root.dep.filehandler.ReadDir(root.metadata.DistDirectory, ".sh")
		if err != nil {
			root.log.Error.Log("Failed to find script files: %v", err)
			fmt.Println("Scripts will not be ran during the deployment")

			root.errors.ScriptsFailed = true
		}

		root.data.scriptFiles = scriptFiles

		// pre script execution
		if len(root.config.Scripts.Pre) > 0 && !root.errors.ScriptsFailed {
			root.log.Debug.Log("Pre script files: %v", root.config.Scripts.Pre)

			root.executeScripts(root.config.Scripts.Pre, root.data.scriptFiles)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := root.config.Admin.InitializeSudo()
		if err != nil {
			root.log.Warn.Log("Failed to authenticate sudo: %v", err)
		}

		root.startAccountCreation(root.dep.usermaker, root.dep.filevault, root.AdminStatus)

		err = root.log.WriteFile()
		if err != nil {
			fmt.Printf("Failed to write to log file: %v\n", err)
		}

		// creating the files found in the search directories, it is flattened.
		searchDirectoryFiles := make([]string, 0)
		for _, searchDir := range root.config.SearchDirectories {
			searchFiles, err := utils.GetFiles(searchDir)
			if err != nil {
				root.log.Warn.Log("Path %s does not exist, skipping path", searchDir)
				continue
			}

			searchDirectoryFiles = append(searchDirectoryFiles, searchFiles...)
		}

		root.log.Debug.Log("File amount: %d | Directories: %v", len(searchDirectoryFiles), root.config.SearchDirectories)

		if len(searchDirectoryFiles) < 1 {
			root.log.Warn.Log("No files found with search directories, all packages will be installed with no checks")
		}

		// used to have len(searchDirectoryFiles) > 0 here, but it doesn't matter just install the files anyways.
		if root.Mount {
			root.log.Info.Log("Searching for DMG files")
			dmgFiles, err := root.dep.filehandler.ReadDir(root.metadata.DistDirectory, ".dmg")
			if err != nil {
				root.log.Error.Log("Failed to search directory: %v", err)
			} else {
				// this requires the use of --include to install properly.
				volumeMounts := root.dep.filehandler.AttachDmgs(dmgFiles)
				if len(volumeMounts) > 0 {
					root.dep.filehandler.AddDmgPackages(volumeMounts, root.metadata.DistDirectory)
					root.dep.filehandler.DetachDmgs(volumeMounts)
				}
			}
		}

		root.startPackageInstallation(root.dep.filehandler, searchDirectoryFiles)

		// app files will automatically get placed into the Applications folder
		appFiles, err := root.dep.filehandler.ReadDir(root.metadata.DistDirectory, ".app")
		if err != nil {
			root.log.Error.Log("Failed to search directory: %v", err)
		}
		if len(appFiles) > 0 {
			applicationDir := "/Applications"
			root.dep.filehandler.CopyFiles(appFiles, applicationDir)
		}

		// inter script execution
		if len(root.config.Scripts.Inter) > 0 && !root.errors.ScriptsFailed {
			root.log.Debug.Log("Inter script files: %v", root.config.Scripts.Inter)

			root.executeScripts(root.config.Scripts.Inter, root.data.scriptFiles)
		}

		err = root.log.WriteFile()
		if err != nil {
			fmt.Printf("Failed to write to log file: %v\n", err)
		}

		err = root.config.Admin.InitializeSudo()
		if err != nil {
			root.log.Warn.Log("Failed to authenticate sudo: %v", err)
		}

		// payload for sending to the server
		// initialized with an empty key, Key gets updated below.
		// used for sending the log over to the server.
		logPayload := requests.NewLogPayload(root.log.GetLogName())
		filevaultPayload := requests.NewFileVaultPayload("")
		if root.config.FileVault {
			fvKey := root.startFileVault(root.dep.filevault)

			filevaultPayload.Key = fvKey
			filevaultPayload.SetBody(root.metadata.SerialTag)
		}
		err = root.log.WriteFile()
		if err != nil {
			fmt.Printf("Failed to write to log file: %v\n", err)
		}

		if root.config.Firewall {
			root.startFirewall(root.dep.firewall)
		}

		// if admin is applied policies, it must be after all the sudo commands.
		// unsure why, but from my testing it fails the filevault command when it was applied
		// prior to running the command.
		if root.config.Admin.ApplyPolicy {
			policyString := root.config.Policy.BuildCommand()

			root.applyPasswordPolicy(policyString, root.config.Admin.Username)
		}

		request := requests.NewRequest(root.log)

		if filevaultPayload.Key != "" {
			root.log.Info.Log("Sending FileVault key to the server")

			err = root.startRequest(filevaultPayload, request, "/api/fv")
			if err != nil {
				root.log.Error.Log("Failed to send to data to server: %v", err)
				fmt.Printf("The key must be saved manually: %s", filevaultPayload.Key)

				err = root.log.WriteFile()
				if err != nil {
					fmt.Printf("Failed to write to log file: %v\n", err)
				}

				root.errors.ServerFailed = true

				return
			}
		} else {
			fvStatus, err := root.dep.filevault.Status()
			if err != nil {
				root.log.Error.Log("Failed to check FileVault status %v", err)
			} else {

				if !fvStatus {
					root.log.Warn.Log("FileVault key failed to generate.")
				}
			}
		}
		if !root.NoSend {
			root.log.Info.Log("Sending log file to the server")

			err = root.log.WriteFile()
			if err != nil {
				fmt.Printf("Failed to write to log file: %v\n", err)
			}

			logPayload.Body = string(root.log.GetContent())
			err = root.startRequest(logPayload, request, "/api/log")
			if err != nil {
				root.log.Error.Log("Failed to send to data to server: %v", err)
			}
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Completed deployment for %s\n", root.metadata.SerialTag)
		fmt.Printf("Log output: %s\n", root.log.GetLogPath())

		// post script execution
		if len(root.config.Scripts.Post) > 0 && !root.errors.ScriptsFailed {
			root.log.Debug.Log("Post script files: %v", root.config.Scripts.Post)

			root.executeScripts(root.config.Scripts.Post, root.data.scriptFiles)
		}

		if root.RemoveFiles {
			if root.errors.ServerFailed {
				choice := ""
				validChoices := "yn"

				validation := func(choice string, validChoices string) bool {
					choice = strings.ToLower(choice)
					validChoices = strings.ToLower(validChoices)

					validArr := strings.Split(validChoices, "")

					return slices.Contains(validArr, choice)
				}

				reader := bufio.NewReader(os.Stdin)

				fmt.Println("The FileVault key failed to send to the server.")
				fmt.Println("You can re-run the binary to try again.")
				for !validation(choice, validChoices) {
					fmt.Print("Remove all deployment files? [y/N]: ")
					choice, _ = reader.ReadString('\n')

					choice = strings.ToLower(choice)
					choice = strings.TrimSpace(choice)

					if choice == "n" || choice == "" {
						return
					} else if choice == "y" {
						break
					} else {
						fmt.Printf("invalid response [%s]\n", choice)
					}
				}
			}

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

// InitializeRoot initializes the flags for rootCmd.
func InitializeRoot() {
	rootCmd.Flags().StringArrayVar(&root.ExcludePackages,
		"exclude", []string{}, "Exclude a package from installing")
	rootCmd.Flags().StringArrayVar(&root.IncludePackages,
		"include", []string{}, "Include a package to install")
	rootCmd.Flags().StringVar(
		&root.PlistPath, "plist", "", "Apply password policies with a plist path")

	rootCmd.Flags().BoolVarP(
		&root.AdminStatus, "admin", "a", false, "Grants admin to the created user")
	rootCmd.Flags().BoolVar(
		&root.RemoveFiles, "remove-files", false, "Remove the deployment files on the device upon successful execution")
	rootCmd.Flags().BoolVarP(
		&root.Verbose, "verbose", "v", false, "Displays the info output to the terminal")
	rootCmd.Flags().BoolVar(
		&root.Debug, "debug", false, "Displays the debug output to the terminal")
	rootCmd.Flags().BoolVar(
		&root.NoSend, "no-send", false, "Prevent the log file from being sent to the server")
	rootCmd.Flags().BoolVar(
		&root.Mount, "mount", false, "Mount all DMG files found inside the distribution folder")
}

// accountCreation starts the account making process.
//
// User is used to call and start the account creation process.
func (r *RootData) startAccountCreation(user *core.UserMaker, filevault *core.FileVault, adminStatus bool) {
	if len(r.config.Accounts) < 1 {
		r.log.Warn.Log("No account information given in YAML file")
		return
	}

	for key := range r.config.Accounts {
		currAccount := r.config.Accounts[key]

		accountName, err := user.CreateAccount(&currAccount, adminStatus)
		if err != nil {
			// if user creation is skipped then dont log the error
			if !strings.Contains(err.Error(), "skipped") {
				logMsg := fmt.Sprintf("Error making user: %v", err)
				r.log.Error.Log(logMsg)
			}

			continue
		}

		err = filevault.AddSecureToken(accountName, currAccount.Password)
		if err != nil {
			r.log.Error.Log("Failed to add user to secure token, manual interaction needed")

			// do not skip this process if secure token fails
			err = user.DeleteAccount(accountName)
			if err != nil {
				r.log.Error.Log("Failed to run user removal command, manual deletion needed: %v", err)

				continue
			}
		}

		if currAccount.ApplyPolicy {
			policyString := r.config.Policy.BuildCommand()

			root.applyPasswordPolicy(policyString, accountName)
		}
	}
}

// startPackageInstallation begins the package installation process.
func (r *RootData) startPackageInstallation(handler *core.FileHandler, searchDirectoryFiles []string) {
	err := handler.InstallRosetta()
	if err != nil {
		r.log.Error.Logf("Failed to install Rosetta: %v", err)
		return
	}

	// removing packages take precedent.
	handler.AddPackages(r.IncludePackages)
	handler.RemovePackages(r.ExcludePackages)

	r.log.Info.Log("Searching for packages")
	packages, err := handler.ReadDir(r.metadata.DistDirectory, ".pkg")
	if err != nil {
		r.log.Error.Log("Issue occurred with searching directory %s: %v", r.metadata.DistDirectory, err)
		return
	}

	if len(packages) < 1 {
		r.log.Warn.Log("No packages found")
		return
	}

	handler.InstallPackages(packages, searchDirectoryFiles)
}

// startFileVault begins the FileVault process and returns the generated key.
func (r *RootData) startFileVault(filevault *core.FileVault) string {
	fvKey := ""

	// doesn't matter if it fails, an attempt will always occur.
	fvStatus, err := filevault.Status()
	if err != nil {
		r.log.Warn.Logf("Failed to check FileVault status: %v", err)
	}

	if !fvStatus {
		fvKey = filevault.Enable(r.config.Admin.Username, r.config.Admin.Password)

		if fvKey != "" {
			fmt.Printf("Generated key %s\n", fvKey)
		}
	}

	return fvKey
}

func (r *RootData) startFirewall(firewall *core.Firewall) {
	fwStatus, err := firewall.Status()
	if err != nil {
		firewallErrMsg := strings.TrimSpace(fmt.Sprintf("Failed to execute Firewall script | %v", err))
		r.log.Error.Logf(firewallErrMsg, 3)
	}

	r.log.Debug.Logf("Firewall status: %t", fwStatus)

	if !fwStatus {
		err = firewall.Enable()
		if err != nil {
			r.log.Error.Log("%v", err)
		}
	} else {
		r.log.Info.Log("Firewall is already enabled")
	}
}

// startRequest sends the logs to the server.
func (r *RootData) startRequest(payload requests.Payload, request *requests.Request, endpoint string) error {
	host := r.config.ServerHost + endpoint

	serverStatus, err := request.VerifyConnection(r.config.ServerHost)
	if err != nil {
		r.log.Error.Logf("Error reaching host: %v", err)

		return err
	}

	if serverStatus {
		res, err := request.POSTData(host, payload)
		if err != nil {
			return err
		}

		// filevault errors will be printed out here, not logged.
		fmt.Printf("Server response: %s\n", res.Content)

		if !strings.Contains(res.Status, "success") {
			r.log.Warn.Log("Failed to send payload to server")

			return errors.New("payload failed to send to server")
		}

		r.log.Info.Log("Successfully sent file to server")
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

// applyPasswordPolicy applies the policies on the given user account.
func (r *RootData) applyPasswordPolicy(policyString string, username string) {
	out := ""
	var err error
	// if a plist is given, it takes precendent over the policies defined in the config
	if r.PlistPath == "" {
		r.log.Debug.Log("Policy string: %s | User: %s", policyString, username)
		out, err = root.config.Policy.SetPolicy(policyString, username)
	} else {
		r.log.Debug.Log("plist path: %s | User: %s", r.PlistPath, username)
		out, err = root.config.Policy.SetPolicyPlist(r.PlistPath, username)
	}
	if err != nil {
		r.log.Warn.Log("Failed to add policy to user %s: %v", username, err)

		return
	}

	if !r.config.Policy.ChangeOnLogin {
		r.log.Warn.Log(`Successfully applied policy, but "change_on_login" in the YAML was not set to true`)
		r.log.Warn.Log(
			`Run the command sudo pwpolicy -u '%s' -setpolicy 'newPasswordRequired=1' for the policies to apply`,
			username)
	} else {
		r.log.Info.Log("Successfully applied policy: %s", out)
	}
}

func (r *RootData) executeScripts(executingScripts []string, scriptPaths []string) {
	for _, scriptFile := range executingScripts {
		if scriptFile == "" {
			continue
		}

		out, err := r.dep.filehandler.ExecuteScript(scriptFile, scriptPaths)
		if err != nil {
			r.log.Error.Log("%v, output: %s", err, out)
			continue
		}

		r.log.Info.Log("Output: %s", out)
	}
}
