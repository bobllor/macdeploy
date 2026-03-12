package utils

import (
	"fmt"
	"os"
	"strings"
)

type Metadata struct {
	Home      string
	SerialTag string
	Files     Files
}

type Files struct {
	ZipFile       string
	DistDirectory string
}

// NewMetadata creates a new metadata information for the project.
func NewMetadata(serialTag string, distDirectory string, zipFile string) *Metadata {
	meta := Metadata{
		Home:      os.Getenv("HOME"),
		SerialTag: serialTag,
		Files: Files{
			ZipFile:       zipFile,
			DistDirectory: distDirectory,
		},
	}

	return &meta
}

// ToString returns a string representation of Metadata.
func (m *Metadata) ToString() string {
	sep := "|"
	slc := []string{}

	appf := func(format string, v ...any) {
		slc = append(slc, fmt.Sprintf(format, v...))
	}

	appf("Home='%s'", m.Home)
	appf("Serial tag='%s'", m.SerialTag)
	appf("Zipfile='%s'", m.Files.ZipFile)
	appf("Dist folder='%s'", m.Files.DistDirectory)

	return strings.Join(slc, sep)
}
