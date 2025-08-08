package requests

type FileVaultInfo struct {
	Key       string `json:"key"`
	SerialTag string `json:"serialTag"`
}

type LogInfo struct {
	LogFileName string
	Body        string
}
