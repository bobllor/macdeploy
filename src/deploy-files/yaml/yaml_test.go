package yaml

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bobllor/assert"
	"github.com/bobllor/macdeploy/src/tests"
)

// getConfig returns a test Config struct with all fields
// filled out with default test values.
func getConfig() *Config {
	accounts := map[string]UserInfo{
		"account_one": {
			Username:    "example.one",
			Password:    "ExamplePassword",
			IgnoreAdmin: false,
			ApplyPolicy: false,
		},
		"account_two": {
			Username: "exampletwo",
		},
	}

	packages := map[string][]string{
		"some pkg name.pkg": {"installed file"},
		"pkg_file_no_ext":   {},
	}

	installDirectories := []string{
		"/Applications",
		"/Library/Application Support",
	}

	scripts := ScriptTypes{
		Pre:  []string{"run-before-process.sh"},
		Mid:  []string{"run-during-process.sh"},
		Post: []string{"run-after-process.sh"},
	}

	admin := UserInfo{
		Username: "admin",
		Password: "AdminPassword",
	}

	policy := Policies{
		ChangeOnLogin:  true,
		RequireAlpha:   true,
		ReusePassword:  3,
		RequireNumeric: true,
		MinChars:       5,
		MaxChars:       15,
	}

	config := &Config{
		Accounts:           accounts,
		Packages:           packages,
		InstallDirectories: installDirectories,
		Scripts:            scripts,
		Admin:              admin,
		Policy:             policy,
		ServerHost:         "https://169.254.1.5:5000",
		Firewall:           true,
		FileVault:          true,
		Cleanup:            "warn",
	}

	return config
}

// getEditableConfig returns a map rerepsentation of Config.
//
// This is editable and can use any value for any field.
func getEditableConfig() (map[string]any, error) {
	config := getConfig()

	buf, err := Marshal(config)
	if err != nil {
		return nil, err
	}

	data, err := Unmarshal(buf)
	if err != nil {
		return nil, err
	}

	return data.(map[string]any), nil
}

func TestReadConfig(t *testing.T) {
	config := getConfig()
	// converting back into bytes for reading
	buf, err := Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	config, err = NewConfig(buf)
	if err != nil {
		t.Fatalf("failed to read config buffer: %v", err)
	}

	err = Validate(config)
	if err != nil {
		t.Fatalf("failed to validate config: %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	fake, err := getEditableConfig()
	tests.Checkf(t, err != nil, "failed to read config: %v", err)

	delete(fake, "cleanup")

	buf, err := Marshal(fake)
	tests.Checkf(t, err != nil, "failed to marshal config: %v", err)
	config, err := NewConfig(buf)
	tests.Checkf(t, err != nil, "failed to create new Config: %v", err)

	tests.Checkf(t, config.Cleanup != "warn", "default value for 'cleanup' is not 'none', got %s", config.Cleanup)

	err = Validate(config)
	tests.Checkf(t, err != nil, "failed to validate 'cleanup': %v", err)
}

func TestValidateConfigNormal(t *testing.T) {
	config := getConfig()

	err := Validate(config)
	tests.Checkf(t, err != nil, "failed to validate config: %v", err)
}

func TestValidateFailWrongCleanup(t *testing.T) {
	fake, err := getEditableConfig()
	tests.Checkf(t, err != nil, "failed to read config: %v", err)

	fake["cleanup"] = "wrong value"

	buf, err := Marshal(fake)
	tests.Checkf(t, err != nil, "failed to marshal config: %v", err)
	config, err := NewConfig(buf)
	tests.Checkf(t, err != nil, "failed to create new Config: %v", err)

	err = Validate(config)
	tests.Checkf(t, err == nil, "expected error from validation with key 'Cleanup': %v", fake["cleanup"])
}

func TestValidateConfigFailIncorrectServerHost(t *testing.T) {
	fake, err := getEditableConfig()
	tests.Checkf(t, err != nil, "failed to read config: %v", err)

	fake["server_host"] = 123

	buf, err := Marshal(fake)
	tests.Checkf(t, err != nil, "failed to marshal config: %v", err)
	config, err := NewConfig(buf)
	tests.Checkf(t, err != nil, "failed to create new Config: %v", err)

	err = Validate(config)
	tests.Checkf(t, err == nil, "expected error from validation with key 'ServerHost': %v", fake["server_host"])
}

func TestValidateConfigFailMissingServerHost(t *testing.T) {
	fake, err := getEditableConfig()
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	delete(fake, "server_host")

	buf, err := Marshal(fake)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	config, err := NewConfig(buf)
	if err != nil {
		t.Fatalf("failed to create new Config: %v", err)
	}

	if config.ServerHost != "" {
		t.Fatalf("expected 'ServerHost' to be empty from config, failed to delete: %s", config.ServerHost)
	}

	err = Validate(config)

	if err == nil {
		t.Fatalf("expected error from validation with key 'ServerHost': %v", fake["server_host"])
	}
}

func TestDefaultValue(t *testing.T) {
	fake, err := getEditableConfig()
	tests.Fatal(t, err, fmt.Sprintf("failed to get editable config: %v", err))

	key := "require_numeric"
	baseValue := fake["policies"].(map[string]any)[key]

	delete(fake["policies"].(map[string]any), key)

	buf, err := Marshal(fake)
	tests.Fatal(t, err, fmt.Sprintf("failed to marshal fake config: %v", err))

	config, err := NewConfig(buf)
	tests.Fatal(t, err, fmt.Sprintf("failed to unmarshal YAML config: %v", err))

	if config.Policy.RequireNumeric == baseValue {
		tests.Fatal(t, err, fmt.Sprintf(
			"default value failed with removal of %s: got %v but expected %v",
			key,
			config.Policy.RequireAlpha,
			!baseValue.(bool),
		))
	}
}

func TestEmptyPackages(t *testing.T) {
	config := getConfig()

	baseString := "pkg_file_no_ext"
	found := false
	for key := range config.Packages {
		if key == baseString {
			found = true
		}
	}

	if found != true {
		t.Fatalf("Missing key %s", baseString)
	}

	if len(config.Packages[baseString]) != 0 {
		t.Fatalf("Package %s is not 0", baseString)
	}
}

func TestBuildPolicyCommand(t *testing.T) {
	config := getConfig()

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

	fmt.Println(config.Cleanup)
}

func TestFormatUrlWithSlash(t *testing.T) {
	baseUrl := "https://testing.com"
	url := baseUrl + "/////"

	newUrl := formatUrl(url)

	assert.Equal(t, newUrl, baseUrl)
}
