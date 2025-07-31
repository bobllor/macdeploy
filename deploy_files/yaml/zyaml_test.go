package yaml

import (
	"fmt"
	"testing"
)

func TestYAMLData(t *testing.T) {
	data := ReadYAML("../../config.yaml")

	fmt.Println(data)
}
