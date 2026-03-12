package requests

type LogPayload struct {
	LogFileName string `json:"logFileName"`
	Body        string `json:"body"`
}

// NewLogPayload creates a new LogPayload.
//
// fileName represents the log file name to be used for storing
// on the server.
func NewLogPayload(fileName string) *LogPayload {
	payload := LogPayload{
		LogFileName: fileName,
		Body:        "",
	}

	return &payload
}

func (l *LogPayload) SetBody(content string) {
	l.Body = content
}
