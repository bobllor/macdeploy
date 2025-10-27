package requests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"macos-deployment/deploy-files/logger"
	"net/http"
)

type Payload interface {
	SetBody(string)
}

type Response struct {
	Status  string
	Content string
}

type Request struct {
	client *http.Client
	log    *logger.Log
}

func NewRequest(log *logger.Log) *Request {
	// used to bypass the unverified check due to no CA
	tls := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	request := Request{
		client: &http.Client{Transport: tls},
		log:    log,
	}

	return &request
}

// POSTData sends a JSON POST request to the server.
func (r *Request) POSTData(host string, payload Payload) (*Response, error) {
	jsonStr, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := r.newJSONRequest(host, jsonStr)
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	r.log.Debug.Log("Response: %v", response)

	return &response, nil
}

// VerifyConnection checks for basic connectivity to the host.
//
// A GET request is sent, if any issues occur an error will be returned.
func (r *Request) VerifyConnection(host string) (bool, error) {
	r.log.Debug.Log("Host: %s", host)

	resp, err := r.client.Get(host)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	r.log.Debug.Log("Response status: %s", resp.Status)

	return resp.StatusCode == 200, nil
}

// newJSONRequest creates a new HTTP request object.
func (r *Request) newJSONRequest(url string, jsonStr []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
