package tests

import (
	"macos-deployment/deploy-files/yaml"
	"strings"
	"testing"
)

var baseYaml string = `
accounts:
  account_one:
    username: "EXAMPLE.ONE"
    password: "PASSWORD"
    ignore_admin: true
  account_two:
    apply_policy: true
packages:
  some pkg name.pkg:
    - "installed file"
  another_pkg_file:
    -
search_directories:
  - "/Applications" 
admin:
  apply_policy: true
policies:
  require_alpha: true
  require_numeric: true
  change_on_login: true
  reuse_password: 3
  min_characters: 5
  max_characters: 15
server_host: "https://192.168.1.154:5000"
log_output: "sample/logs"
filevault: true
firewall: true
`

func TestGetConfig(t *testing.T) {
	_, err := yaml.NewConfig([]byte(baseYaml))
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildPolicyCommand(t *testing.T) {
	config, err := yaml.NewConfig([]byte(baseYaml))
	if err != nil {
		t.Fatal(err)
	}

	policyString := config.Policy.BuildCommand()

	expectedPolicies := []string{
		"requiresAlpha=1", "requiresNumeric=1",
		"newPasswordRequired=1", "usingHistory=3",
		"minChars=5", "maxChars=15",
	}

	for _, policy := range expectedPolicies {
		if !strings.Contains(policyString, policy) {
			t.Fatalf("failed to find %s in built policy: %s", policy, policyString)
		}
	}
}
