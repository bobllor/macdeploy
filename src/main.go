package main

import (
	"bytes"
	_ "embed"
	"fmt"
	embedhandler "macos-deployment/config"
	"macos-deployment/deploy-files/cmd"
	"macos-deployment/deploy-files/core"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	requests "macos-deployment/deploy-files/server-requests"
	"macos-deployment/deploy-files/utils"
	"macos-deployment/deploy-files/yaml"
	"os"
	"os/exec"
)

func main() {
	const projectName string = "macos-deployment"
	const distDirectory string = "dist"
	const zipFile string = "deploy.zip"

	// not exiting, just in case mac fails somehow. but there are checks for non-mac devices.
	serialTag, err := utils.GetSerialTag()
	log := logger.NewLog(serialTag)
	if err != nil {
		log.Error.Println("Unable to get serial number: %v", err)
	}

	metadata := utils.NewMetadata(projectName, serialTag, distDirectory, zipFile)

	config, err := yaml.NewConfig(embedhandler.YAMLBytes)
	if err != nil {
		log.Error.Println(fmt.Sprintf("Error parsing YAML config: %v", err))
		os.Exit(1)
	}

	scripts := scripts.NewScript()

	cmd.InitializeRoot(log, config, scripts, metadata)
	cmd.Execute()

	os.Exit(1) // TODO: REMOVE ME LATER

	var logJsonMap = &requests.LogInfo{}
	var fvJsonData = &requests.FileVaultInfo{}

	if config.FileVault {
		startFileVault(fvJsonData)
	}

	if config.Firewall {
		startFirewall()
	}

	status, err := requests.VerifyConnection(config.ServerHost)
	if err != nil {
		logger.Log(fmt.Sprintf("Unable to connect to server: %s", err.Error()), 3)
		logger.Log("Unable to send FileVault key to the server", 4)

		// used to prevent a cleanup even on fail, allows reruns on the binary
		return
	}
	if status {
		filesToRemove := map[string]struct{}{
			utils.Globals.DistDirName: {},
			utils.Globals.ZIPFileName: {},
		}

		if config.AlwaysCleanup {
			utils.RemoveFiles(filesToRemove)
		}

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

	logUrl := config.ServerHost + "/api/log"
	fvUrl := config.ServerHost + "/api/fv"

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

// pkgInstallation begins the package installation process.
//
// It takes a map of an array of strings, the keys representing the file name and the name of the installed
// package is found in the searchDirFilesArr.
// searchDirFilesArr is a map of strings used only for finding if the package files are found to be installed.
// This data is obtained from the search_directories array YAML config.
func pkgInstallation(packagesMap map[string][]string, searchDirFilesArr []string) {
	pkgPath := utils.Globals.DistDirName
	scriptOut, scriptErr := exec.Command("bash", "-c", scripts.FindPackagesScript, pkgPath).Output()

	// turns out there is an invisible element inside the array...
	if len(foundPKGs) < 2 {
		logger.Log("No packages found", 4)
	}

	for pkge, pkgeArr := range packagesMap {
		isInstalled := core.IsInstalled(pkgeArr, searchDirFilesArr, pkge)
		if !isInstalled {
			err := core.InstallPKG(pkge, foundPKGs)
			if err != nil {
				msgBytes := []byte(err.Error())
				upperFirstBytes := bytes.ToUpper(msgBytes[:1])
				msgBytes[0] = upperFirstBytes[0]

				msg := string(msgBytes)

				logger.Log(fmt.Sprintf("Error with package %s: %s", pkge, msg), 3)
			}
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
