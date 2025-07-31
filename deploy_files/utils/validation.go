package utils

import "net/http"

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
