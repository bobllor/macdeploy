package tests

import (
	"macos-deployment/deploy-files/logger"
)

func GetLogger(logDir string) *logger.Log {
	serialTag := "SERIAL_TAG"
	verbose := false

	return logger.NewLog(serialTag, logDir, verbose)
}
