package scripts

import (
	_ "embed"
)

// IMPORTANT: since these are ran after the -c flag with bash, the args start at $0 rather than $1.
// NOTE: next time, read the man page...

//go:embed create_user.sh
var CreateUserScript string

//go:embed enable_filevault.sh
var EnableFileVaultScript string

//go:embed enable_firewall.sh
var EnableFirewallScript string

//go:embed find_pkgs.sh
var FindPackagesScript string

type BashScripts struct {
	CreateUser      string
	EnableFileVault string
	EnableFirewall  string
	FindPackages    string
}

// NewScript generates an embedded struct for scripts created in Bash.
func NewScript() *BashScripts {
	scripts := BashScripts{
		CreateUser:      CreateUserScript,
		EnableFileVault: EnableFileVaultScript,
		EnableFirewall:  EnableFirewallScript,
		FindPackages:    FindPackagesScript,
	}

	return &scripts
}
