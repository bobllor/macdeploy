package requests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bobllor/assert"
	"github.com/bobllor/macdeploy/src/deploy-files/logger"
)

var testSerial = "ABCDEFG"

func TestGetQueryDataNormalNoServer(t *testing.T) {
	mux := http.NewServeMux()
	serv := httptest.NewServer(mux)
	defer serv.Close()

	mux.HandleFunc("GET /api/devices/{device}", testQueryFunc)
	req := NewRequest(logger.NewTestLogger())

	t.Run("Normal With Device", func(t *testing.T) {

		devQ, err := req.GetDeviceKeyInfo(serv.URL, testSerial)
		assert.Nil(t, err)

		assert.Equal(t, len(devQ.Content), 1)
		assert.Equal(t, devQ.Content[0].Name, testSerial)
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
	serial := "SERIALTAG1"

	dres, err := req.GetDeviceKeyInfo("https://127.0.0.1:5000", serial)
	assert.Nil(t, err)
	assert.NotEqual(t, len(dres.Content), 0)
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

	if strings.EqualFold(device, testSerial) {
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
