package logger

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

type Log struct {
	logFilePath  string
	logFileName  string
	logDirectory string
	content      *bytes.Buffer
	Debug        Logger
	Error        Logger
	Info         Logger
	Warn         Logger
}

type Logger struct {
	silent bool
	logger *log.Logger
}

// NewLog creates a Log struct for logging.
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
		content:      buf,
		logFilePath:  logFilePath,
		logFileName:  logFile,
		logDirectory: logDirectory,
		Debug: Logger{
			logger: log.New(buf, "[DEBUG] ", flag),
			silent: verboseDebug,
		},
		Info:  Logger{logger: log.New(buf, "[INFO] ", flag)},
		Error: Logger{logger: log.New(buf, "[ERROR] ", flag)},
		Warn:  Logger{logger: log.New(buf, "[WARNING] ", flag)},
	}

	return &log
}

// Write writes the contents to the log path.
func (l *Log) WriteFile() error {
	err := os.WriteFile(l.logFilePath, l.content.Bytes(), 0o600)
	if err != nil {
		return err
	}

	return nil
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

// GetContent returns the buffer data in bytes.
func (l *Log) GetContent() []byte {
	return l.content.Bytes()
}

// GetLogDirectory returns the path of the log directory.
func (l *Log) GetLogDirectory() string {
	return l.logDirectory
}

// Log prints the output to the terminal and logs the output.
func (l *Logger) Log(msg string, v ...any) {
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}
	l.logger.Println(msg)

	if !l.silent {
		msg = fmt.Sprintf("%s%s", l.logger.Prefix(), msg)
		fmt.Println(msg)
	}
}

func (l *Logger) Logf(format string, v ...any) {
	l.logger.Printf(format, v...)
}

// FormatLogOutput formats the output path of the log, based on if it ends with a
// slash (/), if it ends in a *.log, or neither (directory).
// If ~ is used then it will expand out to the home directory.
//
// Any path ending in .log will only use the parent.
func FormatLogPath(logPath string) string {
	home := os.Getenv("HOME")
	defaultPath := home + "/logs/macdeploy"
	// just in case backslash is used for some reason. windows is not supported.
	// honestly maybe this should throw an error in a config validation
	logPath = strings.TrimSpace(strings.ReplaceAll(logPath, "\\", "/"))
	logPath = formatSpecialLogPath(logPath)

	logDirArr := strings.Split(logPath, "/")

	newLogArr := make([]string, 0)
	// ensures that the leading slash, if it exists, does not get dropped.
	if logDirArr[0] == "" {
		newLogArr = append(newLogArr, "")
	}
	// drops any spaces from the array, handles any amounts of slashes ( // )
	for _, element := range logDirArr {
		if element != "" {
			newLogArr = append(newLogArr, element)
		}
	}

	// drops .log if it exists at the end of the path
	// NOTE: maybe parse log formatting?
	wordIndex := 0
	for i := len(newLogArr) - 1; i > -1; i-- {
		if !strings.Contains(newLogArr[i], ".log") {
			wordIndex = i
			break
		}
	}

	if logPath == "/" {
		fmt.Printf("Root directory / is not allowed, changing log directory to %s\n", defaultPath)
	}

	// default back to home + logs if it fails.
	if len(newLogArr) <= 1 && newLogArr[0] == "" {
		return defaultPath
	}

	logPath = strings.Join(newLogArr[:wordIndex+1], "/")

	// home expansion for tilde, only on the first occurrence.
	tildePos := strings.Index(logPath, "~")
	if tildePos != -1 {
		// only expand if it precedes a slash or if nothing precedes it
		if tildePos+1 < len(logPath) {
			nextChar, _ := utf8.DecodeRune([]byte{logPath[tildePos+1]})

			if string(nextChar) == "/" {
				logPath = strings.Replace(logPath, "~", home, 1)
			}
		} else {
			if tildePos-1 < 0 {
				logPath = strings.Replace(logPath, "~", home, 1)
			}
		}
	}

	newFirstChar := string(logPath[0])
	if newFirstChar != "/" {
		logPath = "./" + logPath
	}

	return logPath
}

// formatSpecialLogPath formats specific log paths given: ".", and "./".
func formatSpecialLogPath(logPath string) string {
	wd, err := os.Getwd()
	if err != nil {
		return os.Getenv("HOME")
	}

	if logPath == "." || logPath == "./" {
		return wd
	}

	return logPath
}

// MkdirAll utilizes MkdirAll to create the all directories for the log file.
func MkdirAll(logDir string, perm os.FileMode) error {
	err := os.MkdirAll(logDir, perm)
	if err != nil {
		return err
	}

	return nil
}
