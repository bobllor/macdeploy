package requests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// POSTData sends a json POST request to the server.
//
// Errors must be handled.
func POSTData(url string, mapData map[string]any) (string, error) {
	// wtf...
	jsonStr, err := json.Marshal(mapData)
	if err != nil {
		return "", err
	}

	req, err := newJSONRequest(url, jsonStr)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func newJSONRequest(url string, jsonStr []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
