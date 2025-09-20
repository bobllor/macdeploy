package logger

import (
	"bytes"
	"fmt"
	"log"
	"time"
)

type Log struct {
	logFilePath string
	content     *bytes.Buffer
	Debug       *log.Logger
	Error       *log.Logger
	Info        *log.Logger
	Warn        *log.Logger
}

// NewLog creates a Log struct for logging.
// This requires the serial tag of the device.
func NewLog(serialTag string) *Log {
	date := time.Now().Format("2006-01-02T15-04-05")

	logFile := fmt.Sprintf("%s.%s.log", date, serialTag)
	logFilePath := fmt.Sprintf("/tmp/%s", logFile)

	buf := bytes.NewBuffer([]byte{})
	flag := log.Ltime | log.Lmsgprefix

	log := Log{
		content:     buf,
		logFilePath: logFilePath,
		Debug:       log.New(buf, "[DEBUG] ", flag),
		Info:        log.New(buf, "[INFO] ", flag),
		Error:       log.New(buf, "[ERROR] ", flag),
		Warn:        log.New(buf, "[WARNING] ", flag),
	}

	return &log
}
