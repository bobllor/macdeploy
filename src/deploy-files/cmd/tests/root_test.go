package cmd

import (
	"strings"
	"testing"

	"github.com/bobllor/macdeploy/src/deploy-files/core"
	"github.com/bobllor/macdeploy/src/tests"
)

/*
This test file is going to mimic flags being used for the binary.
*/

var includeFiles []string = []string{
	"chrome.pkg",
	"AntIvirus.pkg",
}

func TestAddPackageNoCSV(t *testing.T) {
	handler := core.NewFileHandler(tests.TestLogger)

	handler.AddPackages(includeFiles)

	tests.Checkf(t, len(handler.GetPackages()) != 2, "failed to add packages %v", includeFiles)
}

func TestAddPackageCSV(t *testing.T) {
	newIncludeFiles := make([]string, len(includeFiles))

	installFiles := []string{
		"chrome.app,google enterprise.app",
		"av.app",
	}

	copy(newIncludeFiles, includeFiles)

	for i, s := range newIncludeFiles {
		newIncludeFiles[i] = s + "," + installFiles[i]
	}

	handler := core.NewFileHandler(tests.TestLogger)

	handler.AddPackages(newIncludeFiles)
	i := 0

	for dPkg, dVal := range handler.GetAllPackages() {
		pkgFile := strings.ToLower(includeFiles[i])
		installFile := installFiles[i]

		installArr := strings.Split(installFile, ",")

		tests.Checkf(t,
			len(installArr) != len(dVal),
			"installed files %v (%d) does not match baseline %v (%d)",
			dVal,
			len(dVal),
			installArr,
			len(installArr),
		)

		val := strings.Join(dVal, ",")

		tests.Checkf(t, strings.Contains(dPkg, pkgFile) == false, "string %s not found in %s", pkgFile, dPkg)
		tests.Checkf(t, strings.Contains(val, installFile) == false, "string %s not found in %s", installFile, val)

		i += 1
	}
}

func TestAddMapPackages(t *testing.T) {
	// mimics the packages added via config
	configPkgs := make(map[string][]string)

	for _, file := range includeFiles {
		configPkgs[file] = []string{}
	}

	handler := core.NewFileHandler(tests.TestLogger)

	handler.AddMapPackages(configPkgs)
	i := 0

	for key, val := range handler.GetAllPackages() {
		baseKey := strings.ToLower(includeFiles[i])

		tests.Checkf(t, strings.Contains(baseKey, key) == false, "did not find %s in %s", baseKey, key)
		tests.Checkf(t, len(val) != 0, "expected install files for %s to be 0, got %d", key, len(val))

		i += 1
	}
}
