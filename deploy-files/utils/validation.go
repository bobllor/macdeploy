package utils

import (
	"regexp"
)

// ValidateName checks if the name matches the Regular Expression.
// Returns the name argument and a boolean indicating a valid match.
// Valid name formatting: First Last || First.Last || F Last || F.Last.
func ValidateName(name string) bool {
	regex, _ := regexp.Compile(`^([A-Za-z]+)( |\.)([A-Za-z]+)$`)
	nameBytes := []byte(name)

	return regex.Match(nameBytes)
}
