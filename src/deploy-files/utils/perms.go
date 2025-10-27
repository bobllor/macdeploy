package utils

import "os"

type Perms struct {
	Base       os.FileMode
	BaseDir    os.FileMode
	Executable os.FileMode
	Full       os.FileMode
}

// NewPerms returns a struct of default file permissions.
func NewPerms() *Perms {
	perms := Perms{
		Base:       0o644,
		BaseDir:    0o755,
		Executable: 0o744,
		Full:       0o777,
	}

	return &perms
}
