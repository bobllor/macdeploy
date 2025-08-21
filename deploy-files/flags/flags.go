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
var adminStatus = flag.Bool("a", false, "Used to give Admin privileges to the user.")

func GetFlags() *FlagValues {
	flag.Var(&excludePackages, "exclude", "Exclude a package")
	flag.Parse()

	flags := FlagValues{
		AdminStatus:     *adminStatus,
		ExcludePackages: &excludePackages,
	}

	return &flags
}

func (arrFlag *excludeValue) String() string {
	return fmt.Sprintf("%v", *arrFlag)
}

func (arrFlag *excludeValue) Set(value string) error {
	*arrFlag = append(*arrFlag, value)
	return nil
}
