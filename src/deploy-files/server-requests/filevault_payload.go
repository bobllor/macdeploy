package requests

type FileVaultPayload struct {
	Key  string `json:"key"`
	Body string `json:"serialTag"`
}

func NewFileVaultPayload(key string) *FileVaultPayload {
	payload := FileVaultPayload{
		Key:  key,
		Body: "",
	}

	return &payload
}

func (f *FileVaultPayload) SetBody(content string) {
	f.Body = content
}
