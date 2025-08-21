package flags

import (
	"flag"
	"fmt"
)

type FlagValues struct {
	AdminStatus     bool
	ExcludePackages *excludeValue
}

type excludeValue []string

var excludePackages excludeValue
var adminStatus = flag.Bool("a", false, "Gives Admin privileges to the user.")

func GetFlags() *FlagValues {
	flag.Var(&excludePackages, "exclude", "Exclude a package from installing.")
	flag.Parse()

	flags := FlagValues{
		AdminStatus:     *adminStatus,
		ExcludePackages: &excludePackages,
	}

	return &flags
}

func (a *excludeValue) String() string {
	return fmt.Sprintf("%v", *a)
}

func (a *excludeValue) Set(value string) error {
	*a = append(*a, value)
	return nil
}
