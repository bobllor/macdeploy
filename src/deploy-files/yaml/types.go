package yaml

type Config struct {
	Accounts           map[string]User
	Packages           PackageData
	Search_Directories []string
	Admin              User
	Server_Host        string
	FileVault          bool
	Firewall           bool
	Always_Cleanup     bool
}

type User struct {
	User_Name       string
	Password        string
	Ignore_Admin    bool
	Change_Password bool
}

// Represents the configuration option of map key:value Package_Name:[]Installed_Package_File_Name.
type PackageData map[string][]string
