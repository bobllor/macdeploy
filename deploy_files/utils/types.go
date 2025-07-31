package utils

type Config struct {
	Accounts           map[string]map[string]string
	Packages           map[string][]string
	Search_Directories []string
	File_Vault         bool
	Firewall           bool
}

type User struct {
	FullName string
	UserName string
	Password string
}
