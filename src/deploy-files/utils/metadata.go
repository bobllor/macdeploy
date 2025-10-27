package utils

import (
	"os"
)

type Metadata struct {
	Home          string
	ProjectName   string
	SerialTag     string
	ZipFile       string
	DistDirectory string
}

func NewMetadata(projectName string, serialTag string, distDirectory string, zipFile string) *Metadata {
	meta := Metadata{
		Home:          os.Getenv("HOME"),
		ProjectName:   projectName,
		SerialTag:     serialTag,
		ZipFile:       zipFile,
		DistDirectory: distDirectory,
	}

	return &meta
}
