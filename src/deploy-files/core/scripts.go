package core

import (
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"os/exec"
	"strings"
)

// TODO: maybe support parameters for scripts defined in the YAML config (using delimiters with /).
// TODO: add the flags and parsing logic, this is similar to core.pkg adding packages
// NOTE: scripts are ran after user installation and package installation. add to README.

type ScriptHandler struct {
	script *scripts.BashScripts
	log    *logger.Log
}

func NewScriptHandler(log *logger.Log, scripts *scripts.BashScripts) *ScriptHandler {
	scriptHandler := ScriptHandler{
		log:    log,
		script: scripts,
	}

	return &scriptHandler
}

// GetFilePaths retrieves all paths of files located in a directory to search in,
func (s *ScriptHandler) GetFilePaths(searchString string, searchDirectory string) ([]string, error) {
	out, err := exec.Command("bash", "-c", s.script.FindFiles, searchDirectory, searchString).Output()

	s.log.Debug.Log("Search file executed with params: %s | %s", searchDirectory, searchString)

	if err != nil {
		return nil, err
	}

	pathArr := strings.Split(string(out), "\n")

	s.log.Debug.Log("Files found with %s: %v", searchString, pathArr)

	return pathArr, nil
}

// ExecuteScripts runs shell scripts on the device.
// It requires an array of strings of the script path, and an array of strings of the script names to execute.
//
// The executing scripts are defined from the config or through the flag.
// Any errors that occurs will be skipped and logged.
func (s *ScriptHandler) ExecuteScripts(executingScripts []string, scriptPaths []string) {
	for _, execScriptName := range executingScripts {
		execNameLow := strings.ToLower(execScriptName)
		success := false
		fail := false

		for _, scriptPath := range scriptPaths {
			scriptPathLow := strings.ToLower(scriptPath)

			if strings.Contains(scriptPathLow, execNameLow) {
				s.log.Info.Log("Starting execution for %s", execScriptName)

				// NOTE: if the user exits non-zero on their script, this will fail.
				// will need to write that in the README.
				out, err := exec.Command("bash", "-c", scriptPath).Output()
				outMsg := strings.TrimSpace(string(out))
				if err != nil {
					s.log.Error.Log("Got non-zero exit for %s: %v", execScriptName, err)
					s.log.Error.Log("Output: %s", outMsg)
					fail = true
					break
				}

				s.log.Info.Log("Successfully executed %s", execScriptName)
				s.log.Info.Log("Output: %s", outMsg)

				success = true
				break
			}
		}

		if !success || !fail {
			s.log.Warn.Log("Unable to find %s", execScriptName)
		}
	}
}
