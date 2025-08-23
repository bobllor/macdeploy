package requests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
)

// POSTData sends a json POST request to the server.
//
// Errors must be handled.
func POSTData[J LogInfo | FileVaultInfo](url string, mapData *J) (ResponseData, error) {
	jsonStr, err := json.Marshal(mapData)
	if err != nil {
		return ResponseData{}, err
	}

	req, err := newJSONRequest(url, jsonStr)
	if err != nil {
		return ResponseData{}, err
	}

	// needed for the private server, due to client wipes false cannot be done
	tls := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tls}

	resp, err := client.Do(req)
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
	tls := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tls}

	resp, err := client.Get(url)
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
