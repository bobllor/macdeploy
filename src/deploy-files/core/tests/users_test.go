package core

import (
	"log"
	"os"
	"testing"

	"github.com/bobllor/macdeploy/src/deploy-files/core"
	"github.com/bobllor/macdeploy/src/deploy-files/logger"
	"github.com/bobllor/macdeploy/src/deploy-files/scripts"
	"github.com/bobllor/macdeploy/src/deploy-files/utils"
	"github.com/bobllor/macdeploy/src/deploy-files/yaml"
)

func TestUsernameFormatting(t *testing.T) {
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

func TestFailUserNoPassword(t *testing.T) {
	userInfo := yaml.UserInfo{
		Username: "sample",
		Password: "",
	}

	log := logger.NewLogger(log.New(os.Stdout, "", log.Ldate), logger.Ldebug)

	user := core.NewUser(yaml.UserInfo{
		Username: "admin",
		Password: "admin",
	}, scripts.NewScript(), log)

	_, err := user.CreateAccount(&userInfo, false)
	if err == nil {
		t.Fatal("expected password failure")
	}
}

func TestFailAdminNoPassword(t *testing.T) {
	config := &yaml.Config{}

	// expects to fail due to the input terminal requirement
	err := config.Admin.SetPassword()
	if err == nil {
		t.Fatal(err)
	}
}
