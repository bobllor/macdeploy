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
	silent bool
	*log.Logger
}

// NewLog creates a Log struct for logging.
// This requires the serial tag of the device.
func NewLog(serialTag string, logDirectory string, verbose bool) *Log {
	date := time.Now().Format("2006-01-02T15-04-05")

	logFile := fmt.Sprintf("%s.%s.log", date, serialTag)
	logFilePath := fmt.Sprintf("%s/%s", logDirectory, logFile)

	buf := bytes.NewBuffer([]byte{})
	flag := log.Ltime | log.Lmsgprefix

	// debug will always be silent unless verbose is used.
	verboseDebug := true
	if verbose {
		verboseDebug = false
	}

	log := Log{
		content:     buf,
		logFilePath: logFilePath,
		logFileName: logFile,
		Debug: Logger{
			Logger: log.New(buf, "[DEBUG] ", flag),
			silent: verboseDebug,
		},
		Info:  Logger{Logger: log.New(buf, "[INFO] ", flag)},
		Error: Logger{Logger: log.New(buf, "[ERROR] ", flag)},
		Warn:  Logger{Logger: log.New(buf, "[WARNING] ", flag)},
	}

	return &log
}

// GetLogName returns the log file name ending in .log.
// This is not the full file path.
func (l *Log) GetLogName() string {
	return l.logFileName
}

// GetLogPath returns the full path to the log file.
func (l *Log) GetLogPath() string {
	return l.logFilePath
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

// Log prints the output to the terminal and logs the output.
func (l *Logger) Log(msg string, v ...any) {
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}
	l.Logger.Println(msg)

	if !l.silent {
		msg = fmt.Sprintf("%s%s", l.Prefix(), msg)
		fmt.Println(msg)
	}
}

// FormatLogOutput formats the output path of the log, based on if it ends with a
// slash (/), if it ends in a *.log, or neither (directory).
//
// Any outputs ending in *.log will drop the *.log.
func FormatLogOutput(logOutput string) string {
	// just in case backslash is used for some reason.
	// honestly maybe this should throw an error in a config validation
	logOutput = strings.ReplaceAll(logOutput, "\\", "/")

	logDirArr := strings.Split(logOutput, "/")
	logDirArrLen := len(logDirArr)
	lastArrElement := logDirArr[len(logDirArr)-1]

	// drops .log or removes any ending slashes.
	if strings.Contains(lastArrElement, ".log") {
		return strings.Join(logDirArr[:logDirArrLen-1], "/")
	} else if lastArrElement == "" {
		wordIndex := logDirArrLen

		for i := logDirArrLen - 1; i > 0; i-- {
			if logDirArr[i] != "" {
				wordIndex = i
				break
			}
		}

		if wordIndex == logDirArrLen {
			logOutput = strings.Join(logDirArr[:wordIndex], "/")
		} else {
			logOutput = strings.Join(logDirArr[:wordIndex+1], "/")
		}
	}

	return logOutput
}

// MkdirAll utilizes MkdirAll to create the all directories for the log file.
func MkdirAll(logDir string, perm os.FileMode) error {
	err := os.MkdirAll(logDir, perm)
	if err != nil {
		return err
	}

	return nil
}
