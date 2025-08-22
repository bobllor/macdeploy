package yaml

type Config struct {
	Accounts           map[string]User
	Packages           PackageData
	Search_Directories []string
	Admin              User
	Server_Ip          string
	Firewall           bool
}

type User struct {
	User_Name    string
	Password     string
	Ignore_Admin bool
}

// Represents the configuration option of map key:value Package_Name:[]Installed_Package_File_Name.
type PackageData map[string][]string
