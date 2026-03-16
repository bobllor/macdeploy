package cmd

import (
	"reflect"
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

	packages := handler.GetAllPackages()

	for i := range len(installFiles) {
		installFile := installFiles[i]
		includeFile := strings.ToLower(includeFiles[i])

		dVal, ok := packages[includeFile]
		tests.Checkf(t, ok == false, "package %s not found in %v", includeFile, packages)

		dInstallFile := strings.Join(dVal, ",")

		tests.Checkf(t, installFile != dInstallFile, "install file %s does not match baseline %s", dInstallFile, installFile)
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

	packagesMap := handler.GetAllPackages()

	for key, val := range configPkgs {
		key = strings.ToLower(key)
		pkgVal, ok := packagesMap[key]

		tests.Checkf(t, ok == false, "key %s not found in packages %v", val, packagesMap)
		tests.Checkf(t, len(val) != len(pkgVal), "expected %s to be len 0, got %d", key, len(val))
	}
}

func TestAddMapPackagesWithInstallFiles(t *testing.T) {
	// mimics the packages added via config
	configPkgs := make(map[string][]string)

	installFiles := [][]string{
		{"chrome.app", "chrome enterprise.app"},
		{"av.app"},
	}

	for i, file := range includeFiles {
		configPkgs[file] = installFiles[i]
	}

	handler := core.NewFileHandler(tests.TestLogger)

	handler.AddMapPackages(configPkgs)

	parsedPackages := handler.GetAllPackages()

	for bKey, bVal := range configPkgs {
		bKey = strings.ToLower(bKey)

		pVal, ok := parsedPackages[bKey]
		tests.Checkf(t, ok == false, "key %s not found in packages %v", bKey, parsedPackages)

		tests.Checkf(t, reflect.DeepEqual(bVal, pVal) == false, "install files %v does not meet baseline %v", pVal, bVal)
	}
}
