package yaml

type Config struct {
	Accounts            map[string]User
	Packages            PackageData
	Search_Directories  []string
	Admin               User
	Server_Host         string
	FileVault           bool
	Firewall            bool
	Always_Cleanup      bool
	Add_Change_Password bool
}

type User struct {
	User_Name    string
	Password     string
	Ignore_Admin bool
}

// Represents the configuration option of map key:value Package_Name:[]Installed_Package_File_Name.
type PackageData map[string][]string
