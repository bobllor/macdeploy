package utils

type Config struct {
	Accounts           map[string]User
	Packages           map[string][]string
	Search_Directories []string
	Admin_Name         string
	Server_Ip          string
	File_Vault         bool
	Firewall           bool
}

type User struct {
	User_Name    string
	Password     string
	Ignore_Admin bool
}

type Global struct {
	Home        string
	ProjectPath string
	PKGPath     string
	SerialTag   string
}
