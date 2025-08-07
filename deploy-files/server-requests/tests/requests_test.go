package requests

import (
	"io"
	"net/http"
	"testing"
)

func TestConnection(t *testing.T) {
	url := "http://127.0.0.1:5000"

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
