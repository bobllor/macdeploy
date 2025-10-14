package yaml

import (
	"fmt"
	"os/exec"
	"strings"
)

type Policies struct {
	ReusePassword  int  `yaml:"reuse_password"`
	RequireAlpha   bool `yaml:"require_alpha"`
	RequireNumeric bool `yaml:"require_numeric"`
	MinChars       int  `yaml:"min_characters"`
	MaxChars       int  `yaml:"max_characters"`
	ChangeOnLogin  bool `yaml:"change_on_login"`
}

// NOTE: i am using pwpolicy -setpolicy for these and is "deprecated", but the user can
// override these policies and its logic by using --plist "<path/to/plist>".
// expiration dates are not used because pwpolicy has a bug setting date values.
// i'll probably look into plist sometime in the future, but it is complicated and not documented well.

// BuildCommand builds the command to setup for execution.
//
// It returns the policies used in the command, e.g. "pol1=value1 pol2=value2 ...".
func (p *Policies) BuildCommand() string {
	policies := make([]string, 0)

	// boolean policy strings
	changeLogin := "newPasswordRequired=%d"
	requireAlpha := "requiresAlpha=%d"
	requireNumeric := "requiresNumeric=%d"

	if p.ChangeOnLogin {
		policies = append(policies, formatPolicy(changeLogin, boolToInt(p.ChangeOnLogin)))
	}
	if p.RequireAlpha {
		policies = append(policies, formatPolicy(requireAlpha, boolToInt(p.RequireAlpha)))
	}
	if p.RequireNumeric {
		policies = append(policies, formatPolicy(requireNumeric, boolToInt(p.RequireNumeric)))
	}

	// integer policy strings
	maxChars := "maxChars=%d"
	minChars := "minChars=%d"
	history := "usingHistory=%d"

	if p.MaxChars != 0 {
		policies = append(policies, formatPolicy(maxChars, p.MaxChars))
	}
	if p.MinChars != 0 {
		policies = append(policies, formatPolicy(minChars, p.MinChars))
	}

	if p.ReusePassword > 15 {
		p.ReusePassword = 15
	} else if p.ReusePassword > 0 {
		policies = append(policies, formatPolicy(history, p.ReusePassword))
	}

	return strings.Join(policies, " ")
}

// SetPolicy runs the password policy on the given user.
//
// It returns the output of the command, if it fails an error is returned.
func (p *Policies) SetPolicy(command string, user string) (string, error) {
	cmd := fmt.Sprintf("sudo pwpolicy -u '%s' -setpolicy '%s'", user, command)

	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// SetPolicyPlist runs the password policy on the given user using a plist.
//
// It returns the output of the command, if it fails an error is returned.
func (p *Policies) SetPolicyPlist(plistPath string, user string) (string, error) {
	cmd := fmt.Sprintf("sudo pwpolicy -u '%s' -setaccountpolicies '%s'", user, plistPath)

	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// boolToInt returns an integer "0" or "1" based on the truthness of the argument.
func boolToInt(boolean bool) int {
	if boolean {
		return 1
	}

	return 0
}

// formatPolicy creates the policy string.
func formatPolicy(policy string, value ...any) string {
	return fmt.Sprintf(policy, value...)
}
