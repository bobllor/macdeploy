package logger

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Log struct {
	logFilePath string
	logFileName string
	content     *bytes.Buffer
	Debug       Logger
	Error       Logger
	Info        Logger
	Warn        Logger
}

type Logger struct {
	*log.Logger
}

// NewLog creates a Log struct for logging.
// This requires the serial tag of the device.
func NewLog(serialTag string, logDirectory string) *Log {
	date := time.Now().Format("2006-01-02T15-04-05")

	// just in case backslash is used for some reason.
	// honestly maybe this should throw an error in a config validation
	logDirectory = strings.ReplaceAll(logDirectory, "\\", "/")

	logDirLen := len(logDirectory)
	logDirArr := strings.Split(logDirectory, "/")

	slashFinder := logDirLen
	for i := logDirLen - 1; i > 0; i-- {
		if logDirArr[i] != "" {
			slashFinder = i
			break
		}
	}

	logDirectory = strings.Join(logDirArr[:slashFinder+1], "/")

	logFile := fmt.Sprintf("%s.%s.log", date, serialTag)
	logFilePath := fmt.Sprintf("%s/%s", logDirectory, logFile)

	buf := bytes.NewBuffer([]byte{})
	flag := log.Ltime | log.Lmsgprefix

	log := Log{
		content:     buf,
		logFilePath: logFilePath,
		logFileName: logFile,
		Debug:       Logger{Logger: log.New(buf, "[DEBUG] ", flag)},
		Info:        Logger{Logger: log.New(buf, "[INFO] ", flag)},
		Error:       Logger{Logger: log.New(buf, "[ERROR] ", flag)},
		Warn:        Logger{Logger: log.New(buf, "[WARNING] ", flag)},
	}

	return &log
}

// GetLogName returns the log file name ending in .log.
// This is not the full file path.
func (l *Log) GetLogName() string {
	return l.logFileName
}

// Write writes the contents to the log file.
func (l *Log) WriteFile() error {
	err := os.WriteFile(l.logFilePath, l.content.Bytes(), 0o600)
	if err != nil {
		return err
	}

	return nil
}

// GetContent returns the buffer data in bytes.
func (l *Log) GetContent() []byte {
	return l.content.Bytes()
}

// SilentPrintln writes to the buffer without printing to the terminal.
func (l *Logger) SilentPrintln(msg string) {
	msg = fmt.Sprintf("%s%s\n", l.Prefix(), msg)

	_, _ = l.Writer().Write([]byte(msg))
}
