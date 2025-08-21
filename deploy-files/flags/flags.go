package flags

import (
	"flag"
	"fmt"
)

type FlagValues struct {
	AdminStatus     bool
	ExcludePackages *arrayValue
	IncludePackages *arrayValue
}

type arrayValue []string

var excludePackages arrayValue
var includePackages arrayValue
var adminStatus = flag.Bool("a", false, "Gives Admin privileges to the user.")

func GetFlags() *FlagValues {
	flag.Var(&excludePackages, "exclude", "Exclude a package from installing.")
	flag.Var(&includePackages, "include", "Include a package to install.")
	flag.Parse()

	flags := FlagValues{
		AdminStatus:     *adminStatus,
		ExcludePackages: &excludePackages,
		IncludePackages: &includePackages,
	}

	return &flags
}

func (a *arrayValue) String() string {
	return fmt.Sprintf("%v", *a)
}

func (a *arrayValue) Set(value string) error {
	*a = append(*a, value)
	return nil
}
