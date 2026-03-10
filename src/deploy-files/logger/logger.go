package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Printer interface {
	Println(v ...any)
	Print(v ...any)
	Printf(format string, v ...any)
}

type Prefix struct {
	debug    string
	info     string
	warn     string
	critical string
	fatal    string
}

type Logger struct {
	log      Printer
	content  []byte
	prefix   Prefix
	logLevel int
}

const (
	Ldebug    = 1
	Linfo     = 2
	Lwarn     = 3
	Lcritical = 4
	Lfatal    = 5
)

// NewLogFile creates a new log file with the current date as the file name.
// It will return the open file, the file name, and an error, if not nil.
// The default file name is "{<fileStr>.}2006-01-02.log".
// The parent directories must exist prior to the function call.
//
// fileStr can be used to add a file name to the log. If no name is preferred,
// then an empty string can be used. This is added to the front of the file log.
//
// This only creates and returns the File, the caller is responsible for closing.
func NewLogFile(fileStr string) (*os.File, error) {
	time := time.Now()

	fileName := ""
	if strings.TrimSpace(fileStr) != "" {
		fileName = fileStr + "."
	}

	formatString := "2006-01-02"
	currDate := time.Format(formatString)

	logFileName := fmt.Sprintf("%s%s.log", fileName, currDate)

	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o666)

	return f, err
}

// NewLogger creates a new Logger.
//
// The printer variable is any type that can print.
//
// The logLevel variable is an integer used as a flag for the minimum
// logging level. If this logLevel is not within a range, then it will
// default to FATAL.
func NewLogger(printer Printer, logLevel int) *Logger {
	// default to
	if logLevel < Ldebug || logLevel > Lfatal {
		logLevel = Lfatal
	}

	logger := Logger{
		log: printer,
		prefix: Prefix{
			debug:    "[DEBUG]",
			info:     "[INFO]",
			warn:     "[WARN]",
			critical: "[CRITICAL]",
			fatal:    "[FATAL]",
		},
		logLevel: logLevel,
		content:  make([]byte, 0),
	}

	return &logger
}

// Debug sends a message at the DEBUG level.
func (l *Logger) Debug(v ...any) {
	vMsg := fmt.Sprint(v...)
	msg := fmt.Sprintf("%s %s", l.prefix.debug, vMsg)
	l.content = append(l.content, []byte(msg)...)

	if l.logLevel <= Ldebug {
		l.log.Print(msg)
	}
}

// Debugf sends a message at the DEBUG level with formatting.
func (l *Logger) Debugf(format string, v ...any) {
	vMsg := fmt.Sprintf(format, v...)
	msg := fmt.Sprintf("%s %s", l.prefix.debug, vMsg)
	l.content = append(l.content, []byte(msg)...)

	if l.logLevel <= Ldebug {
		l.log.Print(msg)
	}
}

// Info sends a message at the INFO level.
func (l *Logger) Info(v ...any) {
	vMsg := fmt.Sprint(v...)
	msg := fmt.Sprintf("%s %s", l.prefix.info, vMsg)
	l.content = append(l.content, []byte(msg)...)

	if l.logLevel <= Linfo {
		l.log.Print(msg)
	}
}

// Infof sends a message at the INFO level with formatting.
func (l *Logger) Infof(format string, v ...any) {
	vMsg := fmt.Sprintf(format, v...)
	msg := fmt.Sprintf("%s %s", l.prefix.info, vMsg)
	l.content = append(l.content, []byte(msg)...)

	if l.logLevel <= Linfo {
		l.log.Print(msg)
	}
}

// Warn sends a message at the WARN level.
func (l *Logger) Warn(v ...any) {
	vMsg := fmt.Sprint(v...)
	msg := fmt.Sprintf("%s %s", l.prefix.warn, vMsg)
	l.content = append(l.content, []byte(msg)...)

	if l.logLevel <= Lwarn {
		l.log.Print(msg)
	}
}

// Warnf sends a message at the WARN level with formatting.
func (l *Logger) Warnf(format string, v ...any) {
	vMsg := fmt.Sprintf(format, v...)
	msg := fmt.Sprintf("%s %s", l.prefix.warn, vMsg)
	l.content = append(l.content, []byte(msg)...)

	if l.logLevel <= Lwarn {
		l.log.Print(msg)
	}
}

// Critical sends a message at the CRITICAL level.
func (l *Logger) Critical(v ...any) {
	vMsg := fmt.Sprint(v...)
	msg := fmt.Sprintf("%s %s", l.prefix.critical, vMsg)
	l.content = append(l.content, []byte(msg)...)

	if l.logLevel <= Lcritical {
		l.log.Print(msg)
	}
}

// Criticalf sends a message at the CRITICAL level with formatting.
func (l *Logger) Criticalf(format string, v ...any) {
	vMsg := fmt.Sprintf(format, v...)
	msg := fmt.Sprintf("%s %s", l.prefix.critical, vMsg)
	l.content = append(l.content, []byte(msg)...)

	if l.logLevel <= Lcritical {
		l.log.Print(msg)
	}
}

// Fatal sends a message at the FATAL level.
func (l *Logger) Fatal(v ...any) {
	vMsg := fmt.Sprint(v...)
	msg := fmt.Sprintf("%s %s", l.prefix.fatal, vMsg)
	l.content = append(l.content, []byte(msg)...)

	if l.logLevel <= Lfatal {
		l.log.Print(msg)
	}
}

// Fatalf sends a message at the FATAL level with formatting.
func (l *Logger) Fatalf(format string, v ...any) {
	vMsg := fmt.Sprintf(format, v...)
	msg := fmt.Sprintf("%s %s", l.prefix.fatal, vMsg)
	l.content = append(l.content, []byte(msg)...)

	if l.logLevel <= Lfatal {
		l.log.Print(msg)
	}
}

// GetContent gets the bytes of Logger, it contains
// the print output of Logger.
func (l *Logger) GetContent() []byte {
	return l.content
}
