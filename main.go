package main

import (
	"flag"
	"fmt"
	"macos-deployment/deploy-files/core"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	requests "macos-deployment/deploy-files/server-requests"
	"macos-deployment/deploy-files/utils"
	"macos-deployment/deploy-files/yaml"
	"os"
	"os/exec"
	"strings"
)

var configPath string = "./config.yaml"
var config yaml.Config = yaml.ReadYAML(configPath)

var installTeamViewer = flag.Bool("t", false, "Installs TeamViewer on the device.")
var adminStatus = flag.Bool("a", false, "Used to give Admin privileges to the user.")

var logJsonMap = &requests.LogInfo{}
var fvJsonData = &requests.FileVaultInfo{}

func main() {
	flag.Parse()
	utils.InitializeGlobals()
	logger.NewLog(utils.Globals.SerialTag)

	var accounts map[string]yaml.User = config.Accounts
	accountCreation(accounts)

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
		packagesMap := core.MakePKG(config.Packages, *installTeamViewer)
		pkgInstallation(packagesMap, searchDirFilesArr)
	}

	if config.File_Vault {
		startFileVault()
	}

	if config.Firewall {
		startFirewall()
	}

	status, err := requests.VerifyConnection(config.Server_Ip)
	if err != nil {
		logger.Log(fmt.Sprintf("Unable to connect to server: %s", err.Error()), 3)
		return
	}

	if status {
		sendPOST()
	}
}

// sendPOST sends the FileVault key and log files to the server.
//
// If there are issues with sending data over, it will be skipped.
func sendPOST() {
	// some issue with serial tag, do not send the log in this case.
	if utils.Globals.SerialTag == "UNKNOWN" {
		logger.Log(fmt.Sprintf("Error with serial tag: %s", utils.Globals.SerialTag), 3)
		return
	}

	logUrl := config.Server_Ip + "/api/log"
	fvUrl := config.Server_Ip + "/api/fv"

	// relative path, since it will be in whatever directory the deploy binary is ran in
	logFilePath := fmt.Sprintf("./%s", logger.LogFile)
	logBytes, err := os.ReadFile(logFilePath)
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading log file: %s | path: %s | file name: %s",
			err.Error(), logFilePath, logger.LogFile), 3)
		return
	}

	fvJsonData.SerialTag = utils.Globals.SerialTag

	if fvJsonData.Key != "" && fvJsonData.SerialTag != "UNKNOWN" {
		res, err := requests.POSTData(fvUrl, fvJsonData)
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

	logJsonMap.Body = string(logBytes)
	logJsonMap.LogFileName = logger.LogFile

	if utils.Globals.SerialTag != "UNKNOWN" {
		_, err := requests.POSTData(logUrl, logJsonMap)
		if err != nil {
			logger.Log(fmt.Sprintf("Error sending log to server: %s | Manual interaction needed", err.Error()), 3)
		}
		logger.Log("Sending log file to server", 6)
	} else {
		logger.Log("Unknown serial tag, skipping log transfer process", 4)
	}
}

func accountCreation(accounts map[string]yaml.User) {
	for key := range accounts {
		currAccount := accounts[key]

		core.CreateAccount(currAccount, *adminStatus)
	}
}

func pkgInstallation(packagesMap map[string][]string, searchDirFilesArr []map[string]bool) {
	roseErr := core.InstallRosetta()
	if roseErr != nil {
		logger.Log("Failed to install Rosetta", 3)
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
		isInstalled := core.IsInstalled(pkgeArr, searchDirFilesArr)
		if !isInstalled {
			core.InstallPKG(pkge, foundPKGs)
		}
	}
}

func startFileVault() {
	fvKey := core.EnableFileVault(config.Admin.User_Name, config.Admin.Password)
	if fvKey != "" {
		fvJsonData.Key = fvKey
		keyMsg := fmt.Sprintf("Generated FileVault key %s", fvKey)
		logger.Log(keyMsg, 6)
	} else {
		fvJsonData.Key = ""
	}
}

func startFirewall() {
	core.EnableFireWall()
}
