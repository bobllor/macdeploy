package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var LogFile string
var LogFilePath string

var WarningLevels map[int]string = map[int]string{
	0: "EMERGENCY",
	1: "ALERT",
	2: "CRITICAL",
	3: "ERROR",
	4: "WARNING",
	5: "NOTIFICATION",
	6: "INFO",
	7: "DEBUG",
}

func NewLog(serialTag string) {
	date := time.Now().Format("01-02T15-04-05")

	LogFile = fmt.Sprintf("%s.%s.log", date, serialTag)
	LogFilePath = fmt.Sprintf("/tmp/%s", LogFile)
}

// Log creates and writes to the log file.
//
// There are 7 levels:
//   - 0: EMERGENCY
//   - 1: ALERT
//   - 2: CRITICAL
//   - 3: ERROR
//   - 4: WARNING
//   - 5: NOTIFICATION
//   - 6: INFO
//   - 7: DEBUG
func Log(msg string, level int) {
	file, err := os.OpenFile(LogFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0744)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	log.SetOutput(file)
	log.SetFlags(log.Ltime | log.Lmsgprefix)

	prefix := fmt.Sprintf("[%s] ", WarningLevels[level])
	log.SetPrefix(prefix)

	// lazy way to display info to the user.
	if level != 7 {
		fmt.Println(msg)
	}
	log.Println(msg)
}
