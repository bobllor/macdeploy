package yaml

import (
	"fmt"
	"macos-deployment/deploy-files/yaml"
	"testing"
)

func TestYAMLData(t *testing.T) {
	data := yaml.ReadYAML("../../config.yaml")

	fmt.Println(data)
}
