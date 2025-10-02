package tests

import (
	"macos-deployment/deploy-files/core"
	"math/rand"
	"os"
	"strconv"
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

	handler := core.NewFileHandler(packagesToInstall, logger.Log)

	handler.AddPackages(packagesToAdd)

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
	for _, pkg := range handler.GetPackages() {
		if _, ok := loweredPackages[pkg]; !ok {
			t.Errorf("value %s does not exist", pkg)
		}
	}
}

func TestAddPackages(t *testing.T) {
	log := GetLogger(t)
	handler := core.NewFileHandler(packagesToInstall, log.Log)

	handler.AddPackages(packagesToAdd)

	packages := handler.GetPackages()
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
	handler := core.NewFileHandler(packagesToInstall, logger.Log)

	expectedLength := len(append(packagesToAdd, handler.GetPackages()...)) - 1

	handler.AddPackages(packagesToAdd)

	randomSelection := strings.ToLower(packagesToAdd[rand.Intn(len(packagesToAdd))])

	handler.RemovePackages([]string{randomSelection})

	newLen := len(handler.GetPackages())

	if newLen != expectedLength {
		t.Errorf("failed to remove package, got %d instead of %d", newLen, expectedLength)
	}
}

func TestInstalledPackages(t *testing.T) {
	log := GetLogger(t)
	handler := core.NewFileHandler(packagesToInstall, log.Log)

	alreadyInstalledCount := 0

	for pkg, installedNames := range packagesToInstall {
		if handler.IsInstalled(installedNames, strings.ToLower(pkg), searchDirectoryFiles) {
			alreadyInstalledCount += 1
		}
	}

	if alreadyInstalledCount != baseLenPkgInstall {
		t.Error("packages failed to install")
	}
}

func TestInstallPackages(t *testing.T) {
	log := GetLogger(t)
	handler := core.NewFileHandler(packagesToInstall, log.Log)

	handler.AddPackages(packagesToAdd)

	installedCount := 0
	expectedLen := len(packagesToAdd)

	for pkg, installedNames := range handler.GetAllPackages() {
		isInstalled := handler.IsInstalled(installedNames, strings.ToLower(pkg), searchDirectoryFiles)
		if !isInstalled {
			installedCount += 1

			err := os.WriteFile(strings.ReplaceAll(log.ProjectDirectory+"/"+pkg, ".pkg", ""), []byte{}, 0o644)
			if err != nil {
				continue
			}
		}
	}

	if installedCount != expectedLen {
		t.Errorf("packages failed to install, %d != %d", installedCount, baseLenPkgInstall)
	}

	files, err := os.ReadDir(log.ProjectDirectory)
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
		err := os.WriteFile(log.ProjectDirectory+"/"+dmgFile, []byte{}, 0o744)
		// hmm...
		if err != nil {
			t.Error(err)
		}
	}

	dmg := core.NewFileHandler(packagesToInstall, log.Log)

	dmgFiles, err := dmg.ReadDir(log.ProjectDirectory, ".dmg")
	if err != nil {
		t.Errorf("failed to read directory: %v", err)
	}

	newLen := len(dmgFiles)

	if newLen != baseLenDmg {
		t.Fatalf("got %d, did not match the baseline %d", newLen, baseLenDmg)
	}
}

func TestCopyApp(t *testing.T) {
	log := GetLogger(t)

	handler := core.NewFileHandler(packagesToInstall, log.Log)

	appBundle := "a program bundle.app"
	appDirectory := "Applications"

	appBundleContentDir := "contents"
	directories := []string{appBundle, appBundle + "/" + appBundleContentDir}

	for _, dir := range directories {
		err := log.Mkdir(dir)
		CheckError(err, t)
	}

	err := log.WriteFile(appBundle+"/"+"file.ini", []byte("text goes here"))
	CheckError(err, t)

	files, err := handler.ReadDir(log.ProjectDirectory, ".app")
	CheckError(err, t)

	handler.CopyFiles(files, log.ProjectDirectory+"/"+appDirectory)

	files, err = handler.ReadDir(log.ProjectDirectory+"/"+appDirectory, ".app")
	CheckError(err, t)

	if len(files) != 1 {
		t.Fatal("failed to read directory")
	}

	files, err = handler.ReadDir(log.ProjectDirectory+"/"+appDirectory, ".ini")
	CheckError(err, t)

	for _, file := range files {
		out, err := os.ReadFile(file)
		CheckError(err, t)

		if string(out) == "" {
			t.Fatal("failed to copy and write files")
		}
	}
}

func TestCopyFile(t *testing.T) {
	log := GetLogger(t)

	handler := core.NewFileHandler(packagesToInstall, log.Log)

	newTestDir := "test-dir"

	err := log.Mkdir(newTestDir)
	CheckError(err, t)

	fileNames := []string{
		"test.txt", "sample.txt", "ok.txt",
	}
	textContent := "some text here"

	for i, fileName := range fileNames {
		err := log.WriteFile(newTestDir+"/"+fileName, []byte(textContent+strconv.Itoa(i)))
		CheckError(err, t)
	}

	files, err := handler.ReadDir(log.ProjectDirectory, ".txt")
	CheckError(err, t)

	if len(files) != len(fileNames) {
		t.Fatalf("expected %d got %d", len(fileNames), len(files))
	}

	handler.CopyFiles(files, log.ProjectDirectory+"/"+newTestDir)

	for _, filePath := range files {
		outBytes, err := os.ReadFile(filePath)
		CheckError(err, t)

		if !strings.Contains(string(outBytes), textContent) {
			t.Fatalf("got content %s expected %s", string(outBytes), textContent)
		}
	}
}
