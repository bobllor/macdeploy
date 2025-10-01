package tests

import (
	"macos-deployment/deploy-files/core"
	"math/rand"
	"os"
	"strings"
	"testing"
)

var packagesToAdd = []string{
	"some thing.pkg",
	"AnoTherCaseSenSITIVE.pkg",
	"no pkg here",
}

var packagesToInstall = map[string][]string{
	"TeamViewer.pkg":   {"teamviewer"},
	"Test packAGe.PkG": {"test.package"},
}

var searchDirectoryFiles = []string{
	"teamviewer", "test.package",
	"example 1", "example 2",
}

var baseLenPkgInstall int = len(packagesToInstall)

var testDmgs = []string{
	"test.dmg", "another one.dmg",
}
var baseLenDmg int = len(testDmgs)

func TestArrayLowerCase(t *testing.T) {
	logger := GetLogger(t)

	installer := core.NewInstaller(packagesToInstall, searchDirectoryFiles, logger.Log, script)

	installer.AddPackages(packagesToAdd)

	loweredPackages := make(map[string]struct{}, 0)

	for pkg := range packagesToInstall {
		pkgLow := strings.ToLower(pkg)

		loweredPackages[pkgLow] = struct{}{}
	}
	for _, pkg := range packagesToAdd {
		pkgLow := strings.ToLower(pkg)

		loweredPackages[pkgLow] = struct{}{}
	}

	// should already be lowered here during the constructor and add packages.
	for _, pkg := range installer.GetPackages() {
		if _, ok := loweredPackages[pkg]; !ok {
			t.Errorf("value %s does not exist", pkg)
		}
	}
}

func TestAddPackages(t *testing.T) {
	log := GetLogger(t)
	installer := core.NewInstaller(packagesToInstall, searchDirectoryFiles, log.Log, script)

	installer.AddPackages(packagesToAdd)

	packages := installer.GetPackages()
	newLen := len(packages)

	if newLen != baseLenPkgInstall+len(packagesToAdd) {
		t.Errorf(
			"starting length: %d does not match ending length of add packages: %d",
			baseLenPkgInstall, newLen,
		)
	}
}

func TestRemovePackages(t *testing.T) {
	logger := GetLogger(t)
	installer := core.NewInstaller(packagesToInstall, searchDirectoryFiles, logger.Log, script)

	expectedLength := len(append(packagesToAdd, installer.GetPackages()...)) - 1

	installer.AddPackages(packagesToAdd)

	randomSelection := strings.ToLower(packagesToAdd[rand.Intn(len(packagesToAdd))])

	installer.RemovePackages([]string{randomSelection})

	newLen := len(installer.GetPackages())

	if newLen != expectedLength {
		t.Errorf("failed to remove package, got %d instead of %d", newLen, expectedLength)
	}
}

func TestInstalledPackages(t *testing.T) {
	log := GetLogger(t)
	installer := core.NewInstaller(packagesToInstall, searchDirectoryFiles, log.Log, script)

	alreadyInstalledCount := 0

	for pkg, installedNames := range packagesToInstall {
		if installer.IsInstalled(installedNames, strings.ToLower(pkg)) {
			alreadyInstalledCount += 1
		}
	}

	if alreadyInstalledCount != baseLenPkgInstall {
		t.Error("packages failed to install")
	}
}

func TestInstallPackages(t *testing.T) {
	log := GetLogger(t)
	installer := core.NewInstaller(packagesToInstall, searchDirectoryFiles, log.Log, script)

	installer.AddPackages(packagesToAdd)

	installedCount := 0
	expectedLen := len(packagesToAdd)

	for pkg, installedNames := range installer.GetAllPackages() {
		isInstalled := installer.IsInstalled(installedNames, strings.ToLower(pkg))
		if !isInstalled {
			installedCount += 1

			err := os.WriteFile(strings.ReplaceAll(log.MainDirectory+"/"+pkg, ".pkg", ""), []byte{}, 0o644)
			if err != nil {
				continue
			}
		}
	}

	if installedCount != expectedLen {
		t.Errorf("packages failed to install, %d != %d", installedCount, baseLenPkgInstall)
	}

	files, err := os.ReadDir(log.MainDirectory)
	// just an extra check, to be honest the statement above is good enough.
	if err != nil {
		return
	}

	if installedCount != len(files) {
		t.Errorf("packages failed to write, got packages: %v", files)
	}
}

func TestReadDmg(t *testing.T) {
	log := GetLogger(t)

	for _, dmgFile := range testDmgs {
		err := os.WriteFile(log.MainDirectory+"/"+dmgFile, []byte{}, 0o744)
		// hmm...
		if err != nil {
			t.Error(err)
		}
	}

	dmg := core.NewInstaller(packagesToInstall, searchDirectoryFiles, log.Log, script)

	dmgFiles, err := dmg.ReadDmgDirectory(log.MainDirectory)
	if err != nil {
		t.Errorf("failed to read directory: %v", err)
	}

	// there is an empty string added to the array
	newLen := len(dmgFiles)

	if newLen != baseLenDmg {
		t.Errorf("got %d, did not match the baseline %d", newLen, baseLenDmg)
	}
}
