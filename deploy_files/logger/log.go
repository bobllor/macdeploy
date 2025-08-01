package logger

import (
	"fmt"
	"log"
	"os"
)

var logFile string = "temp.log"

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
	print("hi")
	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0744)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	log.SetOutput(file)
	log.SetFlags(log.Ltime | log.Lshortfile | log.Lmsgprefix)

	prefix := fmt.Sprintf("[%s] ", WarningLevels[level])
	log.SetPrefix(prefix)

	log.Println(msg)
}
