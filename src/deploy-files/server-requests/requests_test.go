package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bobllor/assert"
	"github.com/bobllor/macdeploy/src/deploy-files/logger"
)

// testNoIntegrationSerial is a serial used only for non-integration testing.
var testNoIntegrationSerial = "ABCDEFG"

// NOTE: any tests that has "integration" in it means that the test server
// must be started.
//
// due to the server being a python application, a test folder is created
// when the test server is started. this can be accessed from the root
// repository: 'testroot'. t.TempDir will not work.
// The variables below are the default test variables used for integration.

var testSerial = "SERIALTAG1"
var testKey = "W9Z5-N3KT-Y7MP-L2RX-Q8VH-D4CB"
var testServerHost = "https://127.0.0.1:5000"

func TestGetQueryDataNormalNoServer(t *testing.T) {
	mux := http.NewServeMux()
	serv := httptest.NewServer(mux)
	defer serv.Close()

	mux.HandleFunc("GET /api/devices/{device}", testQueryFunc)
	req := NewRequest(logger.NewTestLogger())

	t.Run("Normal With Device", func(t *testing.T) {

		devQ, err := req.GetDeviceKeyInfo(serv.URL, testNoIntegrationSerial)
		assert.Nil(t, err)

		assert.Equal(t, len(devQ.Content), 1)
		assert.Equal(t, devQ.Content[0].Name, testNoIntegrationSerial)
	})

	t.Run("Normal No Device", func(t *testing.T) {
		devQ, err := req.GetDeviceKeyInfo(serv.URL, "tester123")
		assert.Nil(t, err)

		assert.Equal(t, len(devQ.Content), 0)
	})
}

func TestGetQueryDataFail(t *testing.T) {
	mux := http.NewServeMux()
	serv := httptest.NewServer(mux)
	defer serv.Close()

	url := "/api/devices"

	mux.HandleFunc("GET "+url+"/{device}", testQueryFunc)

	t.Run("Invalid Method", func(t *testing.T) {
		client := http.Client{}

		b := bytes.NewBuffer([]byte{})
		_, err := client.Post(url+"/ABCDE", "text/html", b)
		assert.NotNil(t, err)
	})
}

func TestGetQueryDataIntegration(t *testing.T) {
	log := logger.NewTestLogger()
	req := NewRequest(log)

	dres, err := req.GetDeviceKeyInfo("https://127.0.0.1:5000", testSerial)
	assert.Nil(t, err)
	assert.NotEqual(t, len(dres.Content), 0)
}

func TestGetQueryDataNoDeviceIntegration(t *testing.T) {
	log := logger.NewTestLogger()
	req := NewRequest(log)
	serial := "DoesnotExist"

	dres, err := req.GetDeviceKeyInfo("https://127.0.0.1:5000", serial)
	assert.Nil(t, err)
	assert.Equal(t, len(dres.Content), 0)
	assert.Equal(t, strings.Contains(dres.Message, "not found"), true)
}

func TestSendKeyPayloadIntegration(t *testing.T) {
	log := logger.NewTestLogger()
	req := NewRequest(log)
	root := getProjectRoot(t) + "/testroot/keys"
	// needed due to github actions debugging
	fmt.Println("Debug root:", root)

	t.Run("New Device", func(t *testing.T) {
		pl := NewFileVaultPayload(testKey)
		serial := "SERIALTAG2"
		pl.SetBody(serial)

		res, err := req.POSTData(testServerHost, "/api/fv", pl)
		assert.Nil(t, err)
		assert.True(t, res.Status == "success")
		assert.True(t, strings.Contains(res.Content, testKey))

		_ = os.RemoveAll(root + "/" + serial)
	})

	t.Run("Replace Device With Existing Key", func(t *testing.T) {
		pl := NewFileVaultPayload(testKey)
		serial := "SERIALTAG2"
		pl.SetBody(serial)

		res, err := req.POSTData(testServerHost, "/api/fv", pl)
		assert.Nil(t, err)
		assert.True(t, res.Status == "success")

		res, err = req.POSTData(testServerHost, "/api/fv", pl)
		assert.Nil(t, err)
		assert.True(t, res.Status == "success")
		assert.True(t, strings.Contains(res.Content, "Replaced") || strings.Contains(res.Content, "replaced"))

		_ = os.RemoveAll(root + "/" + serial)
	})

	t.Run("Replace Device With No Existing Key", func(t *testing.T) {
		pl := NewFileVaultPayload(testKey)
		serial := "SERIALTAG3"
		pl.SetBody(serial)

		serialDir := root + "/" + serial

		err := os.MkdirAll(serialDir, 0o777)
		assert.Nil(t, err)

		dir, err := os.Getwd()
		// needed for debugging an issue with github actions
		fmt.Println("Debug current directory:", dir, err, serialDir)

		res, err := req.POSTData(testServerHost, "/api/fv", pl)
		assert.Nil(t, err)
		fmt.Println("Debug:", res)
		assert.True(t, strings.EqualFold(res.Status, "success"))
		assert.TrueAll(t,
			strings.Contains(res.Content, "no key"),
			strings.Contains(res.Content, testKey),
		)

		_ = os.RemoveAll(serialDir)
	})
}

func TestFormatPath(t *testing.T) {
	urls := [][]string{
		{"https://api:5000"},
		{"https://api:5000", "/api/fv"},
		{"http://api:5000"},
		{"http://api:5000", "/api/fv/another/api/long/here"},
		{"https://api:5000/api/fv"},
		{"https://api:5000", "/api/fv", "/long/here"},
	}

	for _, url := range urls {
		assert.Nil(t, ValidateUrl(url...))
	}
}

func TestFormatPathError(t *testing.T) {
	urls := [][]string{
		{""},
		{"https://api:5000", "/api/fv/"},
		{"api:5000/api/fv"},
		{"//fv/api/"},
		{"fv/api/"},
		{"12345"},
		{},
	}

	for _, url := range urls {
		assert.NotNil(t, ValidateUrl(url...))
	}
}

// testQueryFunc is the handler used to test the Request.GetDeviceKeyInfo
// method.
func testQueryFunc(w http.ResponseWriter, r *http.Request) {
	device := r.PathValue("device")

	dq := &DeviceQuery{
		Content:    []DeviceFileData{},
		Status:     "Successful request",
		StatusCode: http.StatusOK,
		Message:    "Some message here",
	}

	if strings.EqualFold(device, testNoIntegrationSerial) {
		dq.Content = append(dq.Content,
			DeviceFileData{
				Name:     device,
				Modified: time.Now(),
				Size:     1234,
			},
		)
	}

	b, err := json.Marshal(dq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// getProjectRoot retrieves the full path to the project repository.
func getProjectRoot(t *testing.T) string {
	dir, err := os.Getwd()
	assert.Nil(t, err)
	projectName := "MacDeploy"

	splitDir := strings.Split(dir, "/")
	newDir := []string{}

	for i := 0; i < len(dir); i++ {
		d := splitDir[i]

		// handles github action runner with duplicate names
		if strings.EqualFold(d, projectName) {
			tempi := i
			for strings.EqualFold(d, projectName) {
				newDir = append(newDir, d)
				tempi += 1
				d = splitDir[tempi]
			}
			break
		}

		newDir = append(newDir, d)
	}

	return strings.Join(newDir, "/")
}
