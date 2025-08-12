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
func POSTData[J LogInfo | FileVaultInfo](url string, mapData *J) (ResponseData, error) {
	// wtf...
	jsonStr, err := json.Marshal(mapData)
	if err != nil {
		return ResponseData{}, err
	}

	req, err := newJSONRequest(url, jsonStr)
	if err != nil {
		return ResponseData{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ResponseData{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResponseData{}, err
	}

	var jsonResponse ResponseData
	json.Unmarshal(body, &jsonResponse)

	return jsonResponse, nil
}

// VerifyConnection checks for basic connectivity to the host.
//
// A GET request is sent, if any issues occur an error will be returned.
func VerifyConnection(url string) (bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

// newJSONRequest creates a new HTTP request object.
func newJSONRequest(url string, jsonStr []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
