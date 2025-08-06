package scripts

import (
	_ "embed"
)

//go:embed create_user.sh
var CreateUserScript string

//go:embed enable_filevault.sh
var EnableFileVaultScript string

//go:embed enable_firewall.sh
var EnableFirewallScript string

//go:embed find_pkgs.sh
var FindPackagesScript string
