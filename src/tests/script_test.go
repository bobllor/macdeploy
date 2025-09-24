package tests

import (
	_ "embed"
	"os/exec"
	"testing"
)

//go:embed test.sh
var TestHello string

func TestScriptOutput(t *testing.T) {
	out, err := exec.Command("bash", "-c", TestHello, "John Smith", "and you?", "okay").Output()
	if err != nil {
		t.Error(err)
	}

	if len(out) < 1 {
		t.Error("bash script failed to generate an output")
	}
}
