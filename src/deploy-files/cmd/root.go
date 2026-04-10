package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	embedhandler "github.com/bobllor/macdeploy/src/config"
	"github.com/bobllor/macdeploy/src/deploy-files/core"
	"github.com/bobllor/macdeploy/src/deploy-files/logger"
	"github.com/bobllor/macdeploy/src/deploy-files/scripts"
	requests "github.com/bobllor/macdeploy/src/deploy-files/server-requests"
	"github.com/bobllor/macdeploy/src/deploy-files/utils"
	"github.com/bobllor/macdeploy/src/deploy-files/yaml"

	"github.com/spf13/cobra"
)

type RootData struct {
	// AdminStatus indicates that the local account should be admin.
	AdminStatus bool

	// Cleanup is used to remove the deployment files after the process is done.
	Cleanup bool

	// Verbose enables INFO level and above logging to stdout.
	Verbose bool

	// Debug enables DEBUG level and above logging to stdout.
	Debug bool

	// NoSend skips sending the log file to the server.
	NoSend bool

	// SkipLocal skips local account creation.
	SkipLocal bool

	// CreateLocal enables the local account creation.
	CreateLocal bool

	// SkipFileVault skips the FileVault process.
	SkipFileVault bool

	// ExcludePackages is a slice of packages to exclude the files defined in the
	// config file.
	ExcludePackages []string

	// IncludePackages is a slice of packages to install. These must be in the 'dist/'
	// directory.
	IncludePackages []string

	// PlistPath is a path to a plist file, used for password policies.
	PlistPath string

	// logFile the logging file name.
	logFile string

	// log is used to log the process.
	log *logger.Logger

	// config holds the YAML data.
	config *yaml.Config

	// script holds the embedded script files to execute commands.
	script *scripts.BashScripts

	// metadata is used to store meta information of the program.
	metadata *utils.Metadata

	// errors is a struct used to flag errors during the deployment process.
	// It flags during the server communication and searching for script files.
	errors errorFlags

	// data is a slice of paths of the script files, if found.
	data varData

	// dep are the main dependencies used for the core process of the deployment.
	dep dependencies

	// perm are file modes for file creation.
	perm *utils.Perms

	// osFile is the log file, used to defer in calls where it must be closed.
	osFile *os.File
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

const (
	distDirectory string = "dist"
	zipFile       string = "deploy.zip"
)

var root RootData

var paddingMsg int = 2

var rootCmd = &cobra.Command{
	Use:   "macdeploy",
	Short: "MacBook deployment tool",
	Long:  `Automated deployment for MacBooks for ITAMs.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		if root.Verbose && root.Debug {
			return fmt.Errorf("--verbose and --debug cannot be used together")
		}
		if root.SkipLocal && root.CreateLocal {
			return fmt.Errorf("--skip-local/-s and --create-local/-c cannot be used together")
		}

		root.initialize(false)

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if root.osFile != nil {
			defer root.osFile.Close()
		}

		root.log.Infof("Deployment started for %s", root.metadata.SerialTag)
		fmt.Printf("Starting deployment for %s\n", root.metadata.SerialTag)

		root.log.Debugf("Metadata data: %s", root.metadata.ToString())

		err := root.config.Admin.InitializeSudo()
		if err != nil {
			root.log.Warn(fmt.Sprintf("Failed to authenticate sudo: %v", err))
		}

		// skip local also skips YAML configured accounts
		// create local will trigger a manual account creation, but will not
		if !root.SkipLocal || root.CreateLocal {
			root.startAccountCreation(root.AdminStatus)
		}

		// creating the files found in the install directories, it is flattened.
		installDirectoryFiles := make([]string, 0)
		for _, searchDir := range root.config.InstallDirectories {
			searchFiles, err := utils.GetFiles(searchDir)
			if err != nil {
				root.log.Warn(fmt.Sprintf("Path %s does not exist, skipping path", searchDir))
				continue
			}

			installDirectoryFiles = append(installDirectoryFiles, searchFiles...)
		}

		root.log.Debugf("File amount: %d | Directories: %v", len(installDirectoryFiles), root.config.InstallDirectories)

		if len(installDirectoryFiles) < 1 {
			srcPkgMsg := "No files found in search directories, packages will always be attempted to isntall"
			root.log.Warn(srcPkgMsg)
			fmt.Println(srcPkgMsg)
		}

		// NOTE: can make the pkg/dmg/app process efficient by searching once.
		// something to note in the future if needed.

		// automatically mount, extract, and dismount DMG files if they exist.
		root.log.Info("Searching for DMG files")
		dmgFiles, err := root.dep.filehandler.ReadDir(root.metadata.Files.DistDirectory, ".dmg")
		if err != nil {
			root.log.Warnf("Failed to search directory: %v", err)
		} else {
			// this requires the use of --include to install properly.
			volumeMounts := root.dep.filehandler.AttachDmgs(dmgFiles)
			if len(volumeMounts) > 0 {
				root.dep.filehandler.AddDmgPackages(volumeMounts, root.metadata.Files.DistDirectory)
				root.dep.filehandler.DetachDmgs(volumeMounts)
			}
		}

		root.startPackageInstallation(root.dep.filehandler, installDirectoryFiles)

		// app files will automatically get placed into the Applications folder
		appFiles, err := root.dep.filehandler.ReadDir(root.metadata.Files.DistDirectory, ".app")
		if err != nil {
			root.log.Warn(fmt.Sprintf("Failed to search directory: %v", err))
		}
		if len(appFiles) > 0 {
			applicationDir := "/Applications"
			root.dep.filehandler.CopyFiles(appFiles, applicationDir)
		}

		// mid deplyoment script execution
		if len(root.config.Scripts.Mid) > 0 && !root.errors.ScriptsFailed {
			fmt.Println("Executing mid-deployment scripts")
			root.log.Debug(fmt.Sprintf("Mid-script files: %v", root.config.Scripts.Mid))

			root.executeScripts(root.config.Scripts.Mid, root.data.scriptFiles)
		}

		err = root.config.Admin.InitializeSudo()
		if err != nil {
			root.log.Warn(fmt.Sprintf("Failed to authenticate sudo: %v", err))
		}

		request := requests.NewRequest(root.log)
		// payload for sending to the server
		// initialized with an empty key, Key gets updated below.
		// used for sending the log over to the server.
		currDate := time.Now().Format("2006-01-02")
		serverLogFile := fmt.Sprintf("%s.%s.log", root.metadata.SerialTag, currDate)
		logPayload := requests.NewLogPayload(serverLogFile)

		if !root.SkipFileVault {
			filevaultPayload := requests.NewFileVaultPayload("")
			if root.config.FileVault {
				fvKey := root.startFileVault(root.dep.filevault, request)

				filevaultPayload.Key = fvKey
				filevaultPayload.SetBody(root.metadata.SerialTag)

				err := root.startRequest(filevaultPayload, request, "/api/fv")
				if err != nil {
					root.log.Warnf("Failed to send payload with FileVault key: %v", err)
					root.warnFileVaultError(filevaultPayload)
				}
			}

		}

		// if admin is applied policies, it must be after all the sudo commands.
		// unsure why, but from my testing it fails the filevault command when it was applied
		// prior to running the command.
		if root.config.Admin.ApplyPolicy {
			policyString := root.config.Policy.BuildCommand()

			root.applyPasswordPolicy(policyString, root.config.Admin.Username)
		}

		if !root.NoSend {
			root.log.Info("Sending log file to the server")

			logPayload.Body = root.log.String()
			err = root.startRequest(logPayload, request, "/api/log")
			if err != nil {
				root.log.Critical(fmt.Sprintf("Failed to send to data to server: %v", err))
			}
		}

		// firewall must be last, all outbound connections are blocked upon activation.
		// fun fact: i forgot i fixed this issue 4 months ago in a bash only script, and brought it back.
		if root.config.Firewall {
			root.startFirewall(root.dep.firewall)
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Completed deployment for %s\n", root.metadata.SerialTag)

		// post script execution
		if len(root.config.Scripts.Post) > 0 && !root.errors.ScriptsFailed {
			fmt.Println("Executing post-deployment scripts")
			root.log.Debug(fmt.Sprintf("Post-script files: %v", root.config.Scripts.Post))

			root.executeScripts(root.config.Scripts.Post, root.data.scriptFiles)
		}

		if root.Cleanup {
			if root.errors.ServerFailed || root.config.Cleanup == "warn" {
				choice := ""
				validChoices := "yn"

				validation := func(choice string, validChoices string) bool {
					choice = strings.ToLower(choice)
					validChoices = strings.ToLower(validChoices)

					validArr := strings.Split(validChoices, "")

					return slices.Contains(validArr, choice)
				}

				reader := bufio.NewReader(os.Stdin)

				if root.errors.ServerFailed {
					fmt.Println("The FileVault key failed to send to the server.")
					fmt.Println("You can re-run the binary to try again.")
				}

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
						fmt.Printf("Invalid response [%s]\n", choice)
					}
				}
			}

			filesToRemove := []string{root.metadata.Files.ZipFile, root.metadata.Files.DistDirectory}

			fmt.Println("Cleaning deployment files...")
			utils.RemoveFiles(filesToRemove)
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
		&root.PlistPath, "pwlist", "", "Apply password policies with a plist")

	rootCmd.Flags().BoolVarP(
		&root.AdminStatus, "admin", "a", false, "Grants admin to local users")
	rootCmd.Flags().BoolVar(
		&root.Cleanup, "cleanup", false, "Remove the deployment files on the device")
	rootCmd.Flags().BoolVarP(
		&root.Verbose, "verbose", "v", false, "Displays the info output to the terminal")
	rootCmd.Flags().BoolVar(
		&root.Debug, "debug", false, "Displays the debug output to the terminal")
	rootCmd.Flags().BoolVar(
		&root.NoSend, "nosend", false, "Do not send the log file to the server")
	rootCmd.Flags().BoolVarP(
		&root.SkipLocal, "skiplocal", "s", false, "Skip the local user creation")
	rootCmd.Flags().BoolVarP(
		&root.CreateLocal, "createlocal", "c", false, "Create a local user")
	rootCmd.Flags().BoolVar(
		&root.SkipFileVault, "nofilevault", false, "Skip the FileVault process",
	)

	rootCmd.MarkFlagsMutuallyExclusive("skiplocal", "createlocal")
	rootCmd.MarkFlagsMutuallyExclusive("debug", "verbose")
}

// startAccountCreation starts the account making process.
// This is used for the YAML accounts and single use accounts.
func (r *RootData) startAccountCreation(adminStatus bool) {
	fmt.Println("Starting account creation")

	r.log.Debugf("YAML accounts amount: %v", len(r.config.Accounts))
	r.log.Debugf("Skip create user: %v | Create user: %v", r.SkipLocal, r.CreateLocal)

	// this takes precedent over the r.config.Accounts.
	if r.CreateLocal {
		account := yaml.UserInfo{}

		accountName := r.accountCreation(&account, adminStatus)
		if accountName != "" {
			r.postAccountCreation(accountName, account.Password, r.config.Policy.ChangeOnLogin)
		}

		return
	}

	for key := range r.config.Accounts {
		currAccount := r.config.Accounts[key]

		accountName := r.accountCreation(&currAccount, adminStatus)
		if accountName != "" {
			r.postAccountCreation(accountName, currAccount.Password, currAccount.ApplyPolicy)
		}
	}
}

// accountCreation starts the account creation process.
//
// It returns the internal username if successful, otherwise it will return an empty string.
func (r *RootData) accountCreation(currAccount *yaml.UserInfo, adminStatus bool) string {
	accountName, err := r.dep.usermaker.CreateAccount(currAccount, adminStatus)
	if err != nil {
		// if user creation is skipped then dont log the error
		if !strings.Contains(err.Error(), "skipped") {
			logMsg := fmt.Sprintf("Error making user: %v", err)
			r.log.Warn(logMsg)

			fmt.Println("Failed to create account")
		}

		return ""
	}

	return accountName
}

// postAccountCreation applies the post account creation policies and secure token.
func (r *RootData) postAccountCreation(accountName string, accountPassword string, applyPolicy bool) {
	fmt.Println("Applying post-account creation workflow")

	err := r.dep.filevault.AddSecureToken(accountName, accountPassword)
	// major error if true.
	if err != nil {
		r.log.Critical("Failed to add user to secure token")

		userSecureTokenString := []string{
			"WARNING",
			"The secure token has failed to apply to the user",
			"The deployment can be restarted or manual user creation may be required",
		}

		secureErrorMsg := utils.FormatBannerString(userSecureTokenString, paddingMsg)

		fmt.Println(secureErrorMsg)

		// do not skip this process if secure token fails
		err = r.dep.usermaker.DeleteAccount(accountName)
		if err != nil {
			r.log.Warn(fmt.Sprintf("Failed to run user removal command, manual deletion needed: %v", err))

			return
		}
	}

	if applyPolicy {
		policyString := r.config.Policy.BuildCommand()

		r.applyPasswordPolicy(policyString, accountName)
	}

	fmt.Printf("User %s successfully created\n", accountName)
}

// startPackageInstallation begins the package installation process.
//
// handler is the FileHandler.
//
// installDirectoryFiles is a slice of strings that contain the files of installation directories.
func (r *RootData) startPackageInstallation(handler *core.FileHandler, installDirectoryFiles []string) {
	fmt.Println("Starting application installation")
	// must be ran prior to installing software, if this fails then
	// software will not install.
	err := handler.InstallRosetta()
	if err != nil {
		r.log.Warn(fmt.Sprintf("Failed to install Rosetta: %v\n", err))
		fmt.Println("Rosetta failed to install, please try again or run 'macdeploy install <file>...'")

		return
	}

	// removing packages take precedent.
	handler.AddPackages(r.IncludePackages)
	handler.RemovePackages(r.ExcludePackages)

	r.log.Debug(handler.PackageString())
	packages, err := handler.ReadDir(r.metadata.Files.DistDirectory, ".pkg")
	if err != nil {
		r.log.Warnf("Issue occurred with searching directory %s: %v", r.metadata.Files.DistDirectory, err)
		return
	}

	if len(packages) < 1 {
		r.log.Warnf("Packages found in %s: %d", r.metadata.Files.DistDirectory, len(packages))
		fmt.Println("No packages found in 'dist', skipping package installation")
		return
	}

	installCount := handler.InstallPackages(packages, installDirectoryFiles)
	msg := fmt.Sprintf("Installed %d/%d files", installCount, len(handler.GetPackages()))

	r.log.Debug(msg)
	fmt.Println(msg)
}

// startFileVault begins the FileVault process and returns the generated key.
func (r *RootData) startFileVault(filevault *core.FileVault, request *requests.Request) string {
	fmt.Println("Starting FileVault process")
	fvKey := ""

	// doesn't matter if it fails, an attempt will always occur.
	fvStatus, err := filevault.Status()
	if err != nil {
		r.log.Warn(fmt.Sprintf("Failed to check FileVault status: %v\n", err))
		fmt.Println("Unable to check FileVault status")
	}

	// UNKNOWN is set during, if one occurs initialization
	if r.metadata.SerialTag != "UNKNOWN" {
		// handles an edge case if FileVault is enabled but there is no entry in the
		// server
		qRes, err := request.GetDeviceKeyInfo(r.config.ServerHost, r.metadata.SerialTag)
		if err != nil {
			r.log.Warnf("Failed to query device information: %v | Response: %v", err, qRes)
		} else {
			r.log.Debugf("Query status code: %d | Query message: %s", qRes.StatusCode, qRes.Message)

			// if len(qRes.Content) == 0/1, then always attempt FileVault process.
			if len(qRes.Content) == 0 || len(qRes.Content) == 1 {
				r.log.Infof("No FileVault entry found for %s", r.metadata.SerialTag)

				if fvStatus {
					filevault.Disable(r.config.Admin.Username, r.config.Admin.Password)
					fvStatus = false
				}
			}
		}
	}

	if !fvStatus {
		fvKey = filevault.Enable(r.config.Admin.Username, r.config.Admin.Password)

		// this is not logged.
		if fvKey != "" {
			fmt.Printf("Generated key %s\n", fvKey)
		}
	}

	return fvKey
}

func (r *RootData) startFirewall(firewall *core.Firewall) {
	fmt.Println("Starting Firewall process")
	fwStatus, err := firewall.Status()
	if err != nil {
		r.log.Warn(fmt.Sprintf("Failed to execute firewall: %v", err))
		fmt.Println("Failed to check firewall status")
	}

	r.log.Debug(fmt.Sprintf("Firewall status: %t\n", fwStatus))

	if !fwStatus {
		err = firewall.Enable()
		if err != nil {
			r.log.Warn(err.Error())
		}
	} else {
		r.log.Info("Firewall is already enabled")
	}
}

// startRequest sends the payload to the server.
func (r *RootData) startRequest(payload requests.Payload, request *requests.Request, endpoint string) error {
	fmt.Println("Starting payload request")
	host := r.config.ServerHost + endpoint

	serverStatus, err := request.VerifyConnection(r.config.ServerHost)
	if err != nil {
		r.log.Critical(fmt.Sprintf("Error reaching host: %v\n", err))

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
			r.log.Critical("Failed to send payload to server")

			return errors.New("payload failed to send to server")
		}

		r.log.Info("Successfully sent file to server")
	} else {
		r.log.Critical("Unable to connect to host, manual interactions needed")
		fmt.Println("Failed to send payload to server")

		return errors.New("unable to connect to host")
	}

	return nil
}

// applyPasswordPolicy applies the policies on the given user account.
func (r *RootData) applyPasswordPolicy(policyString string, username string) {
	fmt.Printf("Starting password policy application for %s\n", username)
	out := ""
	var err error
	// if a plist is given, it takes precendent over the policies defined in the config
	if r.PlistPath == "" {
		r.log.Debug(fmt.Sprintf("Policy string: %s | User: %s", policyString, username))
		out, err = root.config.Policy.SetPolicy(policyString, username)
	} else {
		r.log.Debug(fmt.Sprintf("plist path: %s | User: %s", r.PlistPath, username))
		out, err = root.config.Policy.SetPolicyPlist(r.PlistPath, username)
	}
	if err != nil {
		r.log.Warn(fmt.Sprintf("Failed to add policy to user %s: %v", username, err))

		return
	}

	if !r.config.Policy.ChangeOnLogin {
		r.log.Warn(`Successfully applied policy, but "change_on_login" in the YAML was not set to true`)
		r.log.Warn(
			fmt.Sprintf(
				`Run the command sudo pwpolicy -u '%s' -setpolicy 'newPasswordRequired=1' for the policies to apply`,
				username))
	} else {
		r.log.Info(fmt.Sprintf("Successfully applied policy: %s", out))
	}
}

func (r *RootData) executeScripts(executingScripts []string, scriptPaths []string) {
	for _, scriptFile := range executingScripts {
		if scriptFile == "" {
			continue
		}

		fmt.Printf("Running script: %s\n", scriptFile)
		out, err := r.dep.filehandler.ExecuteScript(scriptFile, scriptPaths)
		scriptOutMsg := fmt.Sprintf("Script %s output: %s", scriptFile, out)
		if err != nil {
			r.log.Warn(fmt.Sprintf("Failed to run %s: %v", scriptFile, err))
			r.log.Info(scriptOutMsg)

			fmt.Printf("Script %s failed to run\n", scriptFile)
			if out != "" {
				fmt.Println(scriptOutMsg)
			}

			continue
		}

		if out != "" {
			r.log.Info(scriptOutMsg)
			fmt.Println(scriptOutMsg)
		}
	}
}

// warnFileVaultError is used to warn the user on the terminal that FileVault has failed.
func (r *RootData) warnFileVaultError(filevaultPayload *requests.FileVaultPayload) {
	if filevaultPayload.Key != "" {
		serverFailWarning := []string{
			"WARNING",
			"The FileVault key failed to send to the server",
			"The deployment can be restarted or manual activation may be required",
			"The key must be saved: %s", filevaultPayload.Key,
		}
		fvServerFailMsg := utils.FormatBannerString(serverFailWarning, paddingMsg)

		fmt.Println("\n" + fvServerFailMsg)
		root.errors.ServerFailed = true
	} else {
		fvFailStrings := []string{
			"WARNING",
			"FileVault failed to activate and the key failed to generate",
			"The deployment can be restarted or manual activation may be required",
		}
		fvFailMsg := utils.FormatBannerString(fvFailStrings, paddingMsg)

		fvStatus, err := root.dep.filevault.Status()
		if err != nil {
			root.log.Warn(fmt.Sprintf("Failed to check FileVault status %v", err))

			fmt.Println("\n" + fvFailMsg)
		} else {
			if !fvStatus {
				root.log.Warn("FileVault key failed to generate")

				fmt.Println("\n" + fvFailMsg)
			}
		}
	}
}

// initialize initializes the data for RootData.
//
// isSubCommand is a flag used to indicate that the method call is used
// for a sub command. This will skip reading most values from the config file,
// the hook lifecycle injection, and some terminal printing if true.
func (r *RootData) initialize(isSubCommand bool) {
	// not exiting, just in case mac fails somehow. but there are checks for non-mac devices.
	serialTag, err := utils.GetSerialTag()
	if err != nil {
		serialTag = "UNKNOWN"
		if !isSubCommand {
			fmt.Printf("Unable to get serial number: %v\n", err)
		}
	}

	perms := utils.NewPerms()
	r.perm = perms

	// if this fails then it will be from relative path.
	currDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current directory, changing to relative path")
		currDir = "."
	}

	metadata := utils.NewMetadata(serialTag, currDir+"/"+distDirectory, currDir+"/"+zipFile)
	scripts := scripts.NewScript()
	config, err := yaml.NewConfig(embedhandler.YAMLBytes)
	if err != nil {
		// TODO: make this a better error message (incorrect keys, required keys missing, etc)
		fmt.Printf("Error parsing YAML configuration, %v\n", err)
		os.Exit(1)
	}

	validateErr := yaml.Validate(config)
	if validateErr != nil {
		fmt.Println(validateErr)
		os.Exit(1)
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

	logPath := "logs/macdeploy"
	defaultLogDir := fmt.Sprintf("%s/%s", metadata.Home, logPath)

	// mkdir needs full permission for some reason.
	// anything other than full will have permissions of 000.
	// full perms assigns it the normal permissions: rwxr-xr-x. which is odd to me.
	err = os.MkdirAll(defaultLogDir, r.perm.Full)
	if err != nil {
		fmt.Printf("Unable to make logging directory: %v\n", err)
	}

	f, err := logger.NewLogFile(fmt.Sprintf("%s/%s", defaultLogDir, "macdeploy"))
	// logger will has a content field, this will contain the logging data.
	if err != nil {
		fmt.Printf("Failed to create log file: %s\n", err.Error())
		f = os.Stdout
	} else {
		// IMPORTANT: this must be closed later and in any other subcommands!
		r.osFile = f
	}

	baseLog := log.New(f, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	logLevel := logger.Lfatal
	if r.Verbose {
		logLevel = logger.Linfo
	} else if r.Debug {
		logLevel = logger.Ldebug
	}

	log := logger.NewLogger(baseLog, logLevel)

	// dependency initializations
	filevault := core.NewFileVault(config.Admin, scripts, log)
	user := core.NewUser(config.Admin, scripts, log)
	handler := core.NewFileHandler(log)
	firewall := core.NewFirewall(log, scripts)

	handler.AddMapPackages(config.Packages)

	r.log = log
	r.config = config
	r.script = scripts
	r.metadata = metadata

	r.dep.usermaker = user
	r.dep.filehandler = handler
	r.dep.firewall = firewall
	r.dep.filevault = filevault

	// script hooks, this is not applicable to sub commands.
	if !isSubCommand {
		// initialized for the lifecycle during pre, install, and post script stages
		scriptFiles, err := r.dep.filehandler.ReadDir(root.metadata.Files.DistDirectory, ".sh")
		if err != nil {
			r.log.Warn(fmt.Sprintf("Failed to find script files: %v", err))
			fmt.Println("Scripts will not be ran during the deployment")

			r.errors.ScriptsFailed = true
		}

		r.data.scriptFiles = scriptFiles

		// pre script execution
		if len(r.config.Scripts.Pre) > 0 && !root.errors.ScriptsFailed {
			fmt.Println("Executing pre-deployment scripts")
			r.log.Debug(fmt.Sprintf("Pre-script files: %v", root.config.Scripts.Pre))

			r.executeScripts(root.config.Scripts.Pre, root.data.scriptFiles)
		}
	}
}
