package logger

import (
	"fmt"
	"log"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/tests"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var logFlag = log.Ldate | log.Ltime

func TestLogFileCreation(t *testing.T) {
	tempDir := t.TempDir()
	tLog := tempDir + "/" + "test-log"

	f, err := logger.NewLogFile(tLog)
	if err != nil {
		t.Fatalf("Failed to create log file %s: %v", tLog, err)
	}

	defer f.Close()

	writer := log.New(f, "", logFlag)

	log := logger.NewLogger(writer, logger.Ldebug)

	logMsg := "This is a test log"
	log.Info(logMsg)

	files, err := filepath.Glob(tempDir + "/*")
	tests.Fatal(t, err, fmt.Sprintf("Failed to read directory %s: %v", tLog, err))

	for _, file := range files {
		content, err := os.ReadFile(file)
		tests.Fatal(t, err, fmt.Sprintf("Failed to read file %s: %v", file, err))

		contentStr := string(content)

		if !strings.Contains(contentStr, "[INFO]") && !strings.Contains(contentStr, logMsg) {
			t.Fatalf("Log file failed to write to log file %s: got content %s", file, contentStr)
		}
	}
}

func TestLogFormattingMethod(t *testing.T) {
	log := logger.NewLogger(log.New(os.Stdout, "", logFlag), logger.Ldebug)

	formatString := "This is a test %s"
	argString := "log"
	log.Criticalf(formatString, argString)

	content := string(log.GetContent())

	if !strings.Contains(content, fmt.Sprintf(formatString, argString)) {
		t.Fatalf("Formatted log failed to match base log format: %s", fmt.Sprintf(formatString, argString))
	}
}
