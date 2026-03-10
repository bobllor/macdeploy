package logger

import (
	"bytes"
	"fmt"
	"io"
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

func TestLogToTerminalStdout(t *testing.T) {
	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", logFlag), logger.Ldebug)

	msg := "This is a test message"

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	log.Debug(msg)
	log.Info(msg)
	log.Warn(msg)
	log.Critical(msg)
	log.Fatalf("ignore me %s", msg)

	out := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		out <- buf.String()
	}()

	w.Close()
	os.Stdout = old

	capturedOut := <-out

	expectedLogLevels := []string{"INFO", "WARN", "DEBUG"}

	for _, level := range expectedLogLevels {
		if !strings.Contains(capturedOut, level) {
			t.Fatalf("unable to find %s in output: %s", level, capturedOut)
		}
	}

	unexpectedLogLevels := []string{"CRITICAL", "FATAL"}

	for _, level := range unexpectedLogLevels {
		if strings.Contains(capturedOut, level) {
			t.Fatalf("found unexpected level %s in output: %s", level, capturedOut)
		}
	}
}

func TestLogToTerminalStderr(t *testing.T) {
	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", logFlag), logger.Lwarn)

	msg := "This is a test message"

	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	log.Debug(msg)
	log.Info(msg)
	log.Warn(msg)
	log.Critical(msg)
	log.Fatalf("ignore me %s", msg)

	out := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		out <- buf.String()
	}()

	w.Close()
	os.Stderr = old

	capturedOut := <-out

	expectedLogLevels := []string{"FATAL", "CRITICAL"}

	for _, level := range expectedLogLevels {
		if !strings.Contains(capturedOut, level) {
			t.Fatalf("unable to find %s in output: %s", level, capturedOut)
		}
	}

	unexpectedLogLevels := []string{"INFO", "WARN", "DEBUG"}

	for _, level := range unexpectedLogLevels {
		if strings.Contains(capturedOut, level) {
			t.Fatalf("found unexpected level %s in output: %s", level, capturedOut)
		}
	}
}
