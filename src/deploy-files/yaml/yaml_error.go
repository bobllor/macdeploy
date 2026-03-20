package yaml

import "fmt"

// configError holds data to handle errors during
// validation errors for Config.
type ConfigError struct {
	errorMap map[string]string
}

const (
	keyNotFoundFormatStr string = "key %s does not exist"
)

// newConfigError creates a new configError.
//
// keys is a slice of strings that are the keys of a struct used for the ConfigError
// map. These are case sensitive.
//
// This initializes an empty string for the map, it will not create
// the error message mapped to each key.
func NewConfigError(keys []string) *ConfigError {
	configKeys := map[string]string{}

	for _, key := range keys {
		configKeys[key] = ""
	}

	c := &ConfigError{
		errorMap: configKeys,
	}

	return c
}

// SetKeyError sets the error string for the mapped key value.
// format is a format string that is expected to have one argument: the YAML value.
//
// If a key does not exist then it will return an error.
func (ce *ConfigError) SetKeyError(key string, format string) error {
	_, ok := ce.errorMap[key]
	if !ok {
		return fmt.Errorf(keyNotFoundFormatStr, key)
	}

	ce.errorMap[key] = format

	return nil
}

// GetKeyError returns the error string of the given key. If the key does not exist,
// then it will return an error.
//
// The error must be handled.
func (ce *ConfigError) GetKeyError(key string) (string, error) {
	msg, ok := ce.errorMap[key]
	if !ok {
		return "", fmt.Errorf(keyNotFoundFormatStr, key)
	}

	return msg, nil
}
