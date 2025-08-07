package scripts

import (
	_ "embed"
	"fmt"
	"os/exec"
	"testing"
)

//go:embed test.sh
var TestHello string

func TestScript(t *testing.T) {
	//fmt.Println(TestHello)
	out, err := exec.Command("sudo", "bash", "-c", TestHello, "John Smith", "and you?", "okay").Output()
	if err != nil {
		val := err.Error()
		panic(val)
	}

	fmt.Println(string(out))
}
