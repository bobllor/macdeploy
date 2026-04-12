package requests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bobllor/macdeploy/src/deploy-files/logger"
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
	log    *logger.Logger
}

func NewRequest(log *logger.Logger) *Request {
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

type DeviceFileData struct {
	Name     string    `json:"name"`
	Modified time.Time `json:"modified"`
	Size     int       `json:"size"`
}
type StatusType string

const (
	StatusTypeError   StatusType = "error"
	StatusTypeSuccess StatusType = "success"
)

type DeviceQuery struct {
	Content    []DeviceFileData `json:"content"`
	Message    string           `json:"message"`
	Status     StatusType       `json:"status"`
	StatusCode int              `json:"status_code"`
}

// GetDeviceKeyInfo sends a GET request to the url with the device tag to retrieve
// the metadata of its stored file on the disk in a DeviceQuery.
//
// The DeviceQuery contains a slice of DeviceFileData in its contents, which
// can be empty, a length of one, or a length of two:
//   - 0: The device used in the query does not have an entry in the server
//   - 1: The device entry exists
//   - 2: The device's FileVault key entry exists
//
// The first entry of the slice will always be the information of the device,
// while the second entry is information about the device's FileVault key,
// if either exists.
//
// The host is expected to the root URL connection to access the server.
func (r *Request) GetDeviceKeyInfo(host string, deviceTag string) (*DeviceQuery, error) {
	url := host + "/api/devices/" + deviceTag

	res, err := r.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to query device (%s): %v", res.Status, res)
	}

	devRes := &DeviceQuery{}

	err = json.NewDecoder(res.Body).Decode(devRes)
	if err != nil {
		return nil, err
	}

	return devRes, nil
}

// POSTData sends a JSON POST request to the server.
// If a trailing/leading slash exists on the host/endpoint respectively or if either is empty,
// an error will be returned.
//
// host is the root connection of the server.
//
// endpoint is the endpoint used to POST to.
//
// payload is any Payload interface.
func (r *Request) POSTData(host string, endpoint string, payload Payload) (*Response, error) {
	vErr := ValidateUrl(host, endpoint)
	if vErr != nil {
		return nil, vErr
	}

	// validated from above
	url := host + endpoint

	jsonStr, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := r.newJSONRequest(url, jsonStr)
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

	r.log.Debugf("POST response status: %v", response.Status)

	return &response, nil
}

// VerifyConnection checks for basic connectivity to the host.
//
// A GET request is sent, if any issues occur an error will be returned.
func (r *Request) VerifyConnection(host string) (bool, error) {
	r.log.Debugf("Host: %s", host)

	resp, err := r.client.Get(host)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	r.log.Debugf("Response status: %s", resp.Status)

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

// ValidateUrl validates the URL strings. It will return an error if it fails validation.
//
// urlStrings must contain at least one argument. The first argument will always be assumed
// to be the host string. Subsequent arguments are always assumed to be endpoints to be added
// onto the host string.
//
// The host string must always start with 'http', and cannot have a trailing slash.
// The endpoint strings must always start with a leading slash, but cannot have a trailing slash.
func ValidateUrl(urlStrings ...string) error {
	if len(urlStrings) == 0 {
		return errors.New("cannot have empty strings")
	}
	// also catches any empty strings
	if !strings.HasPrefix(urlStrings[0], "http") {
		return fmt.Errorf("string '%s' must start with 'http'", urlStrings[0])
	}
	if strings.HasSuffix(urlStrings[0], "/") {
		return fmt.Errorf("string '%s' cannot have a trailing slash", urlStrings[0])
	}

	for _, url := range urlStrings[1:] {
		if len(url) == 0 {
			return fmt.Errorf("string value cannot be an empty")
		}
		if strings.HasSuffix(url, "/") {
			return fmt.Errorf("string '%s' cannot have a trailing slash", url)
		}
		if !strings.HasPrefix(url, "/") {
			return fmt.Errorf("string '%s' must have a leading slash", url)
		}
	}

	return nil
}
