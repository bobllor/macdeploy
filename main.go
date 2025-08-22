package main

import (
	_ "embed"
	"fmt"
	"macos-deployment/deploy-files/core"
	"macos-deployment/deploy-files/flags"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	requests "macos-deployment/deploy-files/server-requests"
	"macos-deployment/deploy-files/utils"
	"macos-deployment/deploy-files/yaml"
	"os"
	"os/exec"
	"strings"
)

//go:embed config.yml
var yamlBytes []byte
var config *yaml.Config = yaml.ReadYAML(yamlBytes)

func main() {
	utils.InitializeGlobals()
	logger.NewLog(utils.Globals.SerialTag)

	logger.Log(fmt.Sprintf("Starting deployment for %s", utils.Globals.SerialTag), 6)

	var flagValues *flags.FlagValues = flags.GetFlags()

	// mutates the config packages
	core.RemovePKG(config.Packages, *flagValues.ExcludePackages)
	core.AddPKG(config.Packages, *flagValues.IncludePackages)

	var accounts *map[string]yaml.User = &config.Accounts
	accountCreation(accounts, flagValues.AdminStatus)

	var logJsonMap = &requests.LogInfo{}
	var fvJsonData = &requests.FileVaultInfo{}

	var searchDirFilesArr []map[string]bool
	for _, searchDir := range config.Search_Directories {
		searchMap, searchErr := utils.GetFileMap(searchDir)
		if searchErr != nil {
			msg := fmt.Sprintf("Path %s does not exist, skipping path", searchDir)
			logger.Log(msg, 4)
			continue
		}

		searchDirFilesArr = append(searchDirFilesArr, searchMap)
	}

	if len(searchDirFilesArr) > 0 {
		pkgInstallation(config.Packages, searchDirFilesArr)
	}

	if config.File_Vault {
		startFileVault(fvJsonData)
	}

	if config.Firewall {
		startFirewall()
	}

	status, err := requests.VerifyConnection(config.Server_Ip)
	if err != nil {
		logger.Log(fmt.Sprintf("Unable to connect to server: %s", err.Error()), 3)
		logger.Log("Unable to send FileVault key to the server", 4)
		return
	}

	// honestly this check is probably not needed.
	if status {
		sendPOST(fvJsonData, logJsonMap)
	}
}

// sendPOST sends the FileVault key and log files to the server. This is the final
// call of the entire script.
//
// It takes a FileVaultInfo pointer and LogInfo pointer for mutation.
//
// If there are issues with sending data to the server, it will be skipped.
func sendPOST(fvData *requests.FileVaultInfo, logData *requests.LogInfo) {
	// some issue with serial tag, do not send the log in this case.
	if utils.Globals.SerialTag == "UNKNOWN" {
		logger.Log(fmt.Sprintf("Error with serial tag: %s", utils.Globals.SerialTag), 3)
		return
	}

	logUrl := config.Server_Ip + "/api/log"
	fvUrl := config.Server_Ip + "/api/fv"

	logBytes, err := os.ReadFile(logger.LogFilePath)
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading log file: %s | path: %s | file name: %s",
			err.Error(), logger.LogFilePath, logger.LogFile), 3)
		return
	}

	fvData.SerialTag = utils.Globals.SerialTag

	if fvData.Key != "" && fvData.SerialTag != "UNKNOWN" {
		res, err := requests.POSTData(fvUrl, fvData)
		if err != nil {
			logger.Log(fmt.Sprintf("Error sending FileVault to server: %s | Manual interaction needed", err.Error()), 3)
		}
		logger.Log("Sending FileVault key to server", 6)

		if res.Status != "success" {
			return
		} else {
			logger.Log(res.Content, 4)
		}
	}

	logData.Body = string(logBytes)
	logData.LogFileName = logger.LogFile

	if utils.Globals.SerialTag != "UNKNOWN" {
		_, err := requests.POSTData(logUrl, logData)
		if err != nil {
			logger.Log(fmt.Sprintf("Error sending log to server: %s | Manual interaction needed", err.Error()), 3)
		}
		logger.Log("Sending log file to server", 6)
	} else {
		logger.Log("Unknown serial tag, skipping log transfer process", 4)
	}
}

// accountCreation starts the account making process.
//
// It takes a map of the User struct from the YAML file.
func accountCreation(accounts *map[string]yaml.User, adminStatus bool) {
	if len(*accounts) < 1 {
		logger.Log("No account information given in YAML file", 4)
	}

	for key := range *accounts {
		currAccount := (*accounts)[key]

		core.CreateAccount(currAccount, adminStatus)
	}
}

// pkgInstallation begins the package installation process.
//
// It takes a map of an array of strings, the keys representing the file name and the name of the installed
// package is found in the searchDirFilesArr.
// searchDirFilesArr is a map of strings used only for finding if the package files are found to be installed.
// This data is obtained from the search_directories array YAML config.
func pkgInstallation(packagesMap map[string][]string, searchDirFilesArr []map[string]bool) {
	roseErr := core.InstallRosetta()
	if roseErr != nil {
		logger.Log("Failed to install Rosetta | Unable to install packages", 3)
		return
	}

	pkgPath := utils.Globals.PKGPath
	scriptOut, scriptErr := exec.Command("bash", "-c", scripts.FindPackagesScript, pkgPath).Output()

	debug := fmt.Sprintf("PKG folder: %s", pkgPath)
	logger.Log(debug, 7)

	if scriptErr != nil {
		scriptErrMsg := "Failed to locate package folder"
		logger.Log(scriptErrMsg, 3)
		return
	}

	foundPKGs := strings.Split(string(scriptOut), "\n")

	logger.Log(fmt.Sprintf("Packages in folder: %v", foundPKGs), 7)
	// turns out there is an invisible element inside the array...
	if len(foundPKGs) < 2 {
		logger.Log("No packages found", 4)
	}

	for pkge, pkgeArr := range packagesMap {
		isInstalled := core.IsInstalled(pkgeArr, &searchDirFilesArr)
		if !isInstalled {
			core.InstallPKG(pkge, foundPKGs)
		}
	}
}

// startFileVault starts the FileVault process.
//
// It generates the key and sets the FileVaultInfo struct's key value.
func startFileVault(jsonData *requests.FileVaultInfo) {
	fvKey := core.EnableFileVault(config.Admin.User_Name, config.Admin.Password)
	if fvKey != "" {
		jsonData.Key = fvKey
		keyMsg := fmt.Sprintf("Generated FileVault key %s", fvKey)
		logger.Log(keyMsg, 6)
	} else {
		jsonData.Key = ""
	}
}

// startFirewall starts the Firewall process.
func startFirewall() {
	core.EnableFireWall()
}
