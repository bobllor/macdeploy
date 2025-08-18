package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	requests "macos-deployment/deploy-files/server-requests"
	"macos-deployment/deploy-files/utils"
	"net/http"
	"os"
	"testing"
)

var url string = "http://192.168.1.154:5000"

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

	return resp, nil
}

func TestConnection(t *testing.T) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("got %d, expected %d", resp.StatusCode, 200)
	}
}

func TestFVPost(t *testing.T) {
	apiUrl := url + "/api/fv"
	sampleData := map[string]string{
		"key":    "CFDS-231S-456S-31LO",
		"serial": "C02NONLULBI01",
	}

	jsonStr, err := json.Marshal(sampleData)
	if err != nil {
		t.Errorf("%v", err)
	}

	resp, err := jsonPost(apiUrl, jsonStr)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var jsonResponse requests.ResponseData
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		t.Errorf("Failed to parse JSON %s", err.Error())
	}

	fmt.Println(jsonResponse.Content)
	fmt.Println(jsonResponse.Status)
}

func TestLogPost(t *testing.T) {
	utils.InitializeGlobals()
	content, err := os.ReadFile(fmt.Sprintf("%s/%s", utils.Globals.ProjectPath, "README.md"))
	if err != nil {
		t.Errorf("unable to read file %s", err.Error())
	}

	sampleData := map[string]string{
		"logFileName": "README.log",
		"body":        string(content),
	}
	url := "http://127.0.0.1:5000/api/log"

	jsonStr, err := json.Marshal(sampleData)
	if err != nil {
		t.Errorf("%v", err.Error())
	}

	resp, err := jsonPost(url, jsonStr)
	if err != nil {
		t.Errorf("%v", err.Error())
	}

	defer resp.Body.Close()
}

func TestJSON(t *testing.T) {
	sampleData := map[string]string{
		"test": "a sentence here",
		"okay": "what tf is wrong wit u!!",
	}

	jsonBytes, err := json.Marshal(sampleData)
	if err != nil {
		panic(jsonBytes)
	}

	if string(jsonBytes) == "" {
		t.Errorf("failed to parse to JSON, check sample data")
	}
}
