package utils

import (
	"fmt"
	"net/http"
	"strings"
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

func TestName(t *testing.T) {
	names := []string{
		"john doe", "Will Smith",
		"lebron.james", "Steven.Curry",
		"j doe", "W Smith",
		"l.james", "S.Curry",
	}

	var fails []string
	var successes []string

	for _, name := range names {
		if !ValidateName(name) {
			fails = append(fails, name)
		} else {
			successes = append(successes, FormatName(name))
		}
	}

	stringNames := strings.Join(names, ", ")

	fmt.Printf("Names: %s\n", stringNames)

	if len(fails) > 0 {
		failedNames := strings.Join(fails, ", ")
		t.Errorf("Failed names: %s\n", failedNames)
	}

	formattedNames := strings.Join(successes, ", ")
	fmt.Printf("Formatted names: %s\n", formattedNames)
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
