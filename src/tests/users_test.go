package tests

import (
	"macos-deployment/deploy-files/core"
	"macos-deployment/deploy-files/scripts"
	"macos-deployment/deploy-files/utils"
	"macos-deployment/deploy-files/yaml"
	"testing"
)

func TestFormatUsername(t *testing.T) {
	names := []string{
		"...john ..-doe",
		"!@#$%^&*()lebron.   !~`:\"?<>,;'|\\{[]]}james!!++==",
		"sold!@ier_   \n\n\t//fro   m-tf??<>:\\.\\2",
		"12345!@#67%$#8*&^(9[]{}0",
		"-...-!!/\\",
		"....a",
		"..-a!!;'",
		".-ab",
	}

	expectedNames := map[string]struct{}{
		"john..-doe":        {},
		"lebron.james":      {},
		"soldier_from-tf.2": {},
		"a1234567890":       {},
		"a-...-":            {},
		"a":                 {},
		"-a":                {}, // this works yes, i tested it on a macbook.
		"-ab":               {},
	}

	for _, name := range names {
		newName := utils.FormatUsername(name)
		if _, ok := expectedNames[newName]; !ok {
			t.Errorf("format failed for name: %s\n", newName)
		}

	}
}

func TestUserNoPassword(t *testing.T) {
	userInfo := yaml.UserInfo{
		Username: "sample",
		Password: "",
	}

	logger := GetLogger(t)

	user := core.NewUser(yaml.UserInfo{
		Username: "admin",
		Password: "admin",
	}, scripts.NewScript(), logger.Log)

	_, err := user.CreateAccount(userInfo, false)
	if err == nil {
		t.Error("expected password failure")
	}
}

func TestAdminNoPassword(t *testing.T) {
	config := &yaml.Config{}

	// expects to fail due to the input terminal requirement
	err := config.SetAdminPassword()
	if err == nil {
		t.Fatal(err)
	}
}
