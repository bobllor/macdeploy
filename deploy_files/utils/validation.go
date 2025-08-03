package utils

import (
	"net/http"
	"regexp"
)

// ValidateServer checks if the server is reachable and returns a 200 if successful.
// If unsuccessful, the matching status code and an error is returned.
func ValidateServer() (int, error) {
	url := "http://10.142.46.165:8000"
	res, resErr := http.Get(url)

	if resErr != nil {
		return res.StatusCode, resErr
	}

	return res.StatusCode, nil
}

// ValidateName checks if the name matches the Regular Expression.
// Returns the name argument and a boolean indicating a valid match.
// Valid name formatting: First Last || First.Last || F Last || F.Last.
func ValidateName(name string) bool {
	regex, _ := regexp.Compile(`^([A-Za-z]+)( |\.)([A-Za-z]+)$`)
	nameBytes := []byte(name)

	return regex.Match(nameBytes)
}
