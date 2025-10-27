package requests

type LogPayload struct {
	LogFileName string `json:"logFileName"`
	Body        string `json:"body"`
}

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
