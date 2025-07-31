package utils

import (
	"net/http"
	"testing"
)

func TestRequest(t *testing.T) {
	var url string = "https://www.google.com"
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("Got error %e", err)
	}

	println(res.StatusCode)
}

/*
func TestGetFileMap(t *testing.T) {
	// contains a correct and incorrect path
	paths := [2]string{Home, "/non-existent/path"}
	for _, path := range paths {
		mapOut, mapErr := GetFileMap(path)
		if mapErr != nil {
			t.Errorf("Invalid path %p\n", mapErr)
		}
		fmt.Printf("Map: %v\n", mapOut)
	}
}
*/
