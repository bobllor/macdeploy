package requests

import (
	"bytes"
	"fmt"
	"io"
	"macos-deployment/deploy-files/utils"
	"net/http"
	"os"
	"testing"
)

var url string = "http://127.0.0.1:5000"

func jsonPost(url string, jsonStr []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return resp, nil
}

func TestConnection(t *testing.T) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	println(string(body))
}

func TestFVPost(t *testing.T) {
	apiUrl := url + "/api/fv"
	var jsonStr = []byte(`{"key": "ABCD-EFG3-LFK5-LO69", "serial": "C02FKYSLOL"}`)

	resp, err := jsonPost(apiUrl, jsonStr)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}

func TestLogPost(t *testing.T) {
	content, err := os.ReadFile(fmt.Sprintf("%s/%s", utils.MainDir, "08-07T23-15-49.UNKNOWN.log"))
	if err != nil {
		panic(err)
	}

	fmt.Println(string(content))
}
