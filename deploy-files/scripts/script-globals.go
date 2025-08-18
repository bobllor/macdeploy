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
