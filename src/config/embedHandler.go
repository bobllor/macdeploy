package embedhandler

import _ "embed"

//go:embed config.yml
var YAMLBytes []byte
