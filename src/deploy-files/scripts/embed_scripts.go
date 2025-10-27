package scripts

import (
	_ "embed"
)

// IMPORTANT: since these are ran after the -c flag with bash, the args start at $0 rather than $1.
// NOTE: next time, read the man page...

//go:embed create_user.sh
var createUserScript string

//go:embed enable_filevault.sh
var enableFileVaultScript string

//go:embed enable_firewall.sh
var enableFirewallScript string

//go:embed find_files.sh
var findFilesScript string

type BashScripts struct {
	CreateUser      string
	EnableFileVault string
	EnableFirewall  string
	FindFiles       string // Takes two arguments: search_dir, ext_type
}

// NewScript generates an embedded struct for scripts created in Bash.
func NewScript() *BashScripts {
	scripts := BashScripts{
		CreateUser:      createUserScript,
		EnableFileVault: enableFileVaultScript,
		EnableFirewall:  enableFirewallScript,
		FindFiles:       findFilesScript,
	}

	return &scripts
}
