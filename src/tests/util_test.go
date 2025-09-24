package tests

import (
	"macos-deployment/deploy-files/utils"
	"testing"
)

func TestFormatUsername(t *testing.T) {
	names := []string{
		"...john ..-doe",
		"!@#$%^&*()lebron.   !~`:\"?<>,;'|\\{[]]}james!!++==",
		"sold!@ier_   //fro   m-tf??<>:\\.\\2",
		"12345!@#67%$#8*&^(9[]{}0",
		"-...-!!/\\",
		"....a",
		"..-a!!;'",
		".-ab",
	}

	expectedNames := map[string]struct{}{
		"john..-doe":        {},
		"lebron.james":      {},
		"soldier_from-tf.2": {},
		"a1234567890":       {},
		"a-...-":            {},
		"a":                 {},
		"-a":                {}, // this works yes, i tested it on a macbook.
		"-ab":               {},
	}

	for _, name := range names {
		newName := utils.FormatUsername(name)
		if _, ok := expectedNames[newName]; !ok {
			t.Errorf("format failed for name: %s\n", newName)
		}

	}
}
