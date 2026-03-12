package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bobllor/macdeploy/src/deploy-files/logger"
	"github.com/bobllor/macdeploy/src/tests"
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

func TestLogContent(t *testing.T) {
	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", logFlag), logger.Ldebug)

	logMsgs := []string{}

	for i := range 5 {
		logMsgs = append(logMsgs, "test"+strconv.Itoa(i))
	}

	logFunctions := []func(v ...any){
		log.Debug,
		log.Info,
		log.Warn,
		log.Critical,
		log.Fatal,
	}

	for i, fn := range logFunctions {
		fn(logMsgs[i])
	}

	content := log.GetContent()

	if len(content) != len(logMsgs) {
		t.Fatalf("failed to write log messages to content, got %d instead of %d", len(content), len(logMsgs))
	}

	for i := range len(content) {
		baseMsg := logMsgs[i]
		contentMsg := content[i]

		if !strings.Contains(contentMsg, baseMsg) {
			t.Fatalf("logged content message '%s' does not contain base string '%s'", contentMsg, baseMsg)
		}
	}
}

func TestLogContentString(t *testing.T) {
	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", logFlag), logger.Ldebug)

	logMsgs := []string{}

	for i := range 5 {
		logMsgs = append(logMsgs, "test"+strconv.Itoa(i))
	}

	logFunctions := []func(v ...any){
		log.Debug,
		log.Info,
		log.Warn,
		log.Critical,
		log.Fatal,
	}

	for i, fn := range logFunctions {
		fn(logMsgs[i])
	}

	content := log.GetContentString()

	contentArr := strings.Split(content, "\n")

	if len(contentArr) != len(logMsgs) {
		t.Fatalf("logged contents does not match baseline log messages: %d != %d", len(contentArr), len(logMsgs))
	}

	for i := range len(contentArr) {
		baseMsg := logMsgs[i]
		contentMsg := contentArr[i]

		if !strings.Contains(contentMsg, baseMsg) {
			t.Fatalf("logged content message '%s' does not contain base string '%s'", contentMsg, baseMsg)
		}
	}
}

func TestLogFormattingMethod(t *testing.T) {
	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", logFlag), logger.Ldebug)

	formatString := "This is a test %s"
	argString := "log"
	log.Criticalf(formatString, argString)

	content := strings.Join(log.GetContent(), "\n")

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
