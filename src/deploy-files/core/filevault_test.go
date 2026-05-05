package core

import (
	"fmt"
	"testing"

	"github.com/bobllor/assert"
	"github.com/bobllor/macdeploy/src/deploy-files/logger"
	"github.com/bobllor/macdeploy/src/deploy-files/yaml"
)

func TestKeyParse(t *testing.T) {
	key := "12345-key-here"
	out := fmt.Sprintf("Output = '%s'", key)
	f := NewFileVault(yaml.UserInfo{}, nil, logger.NewTestLogger())

	t.Run("Normal output", func(t *testing.T) {
		newKey := f.parseKey(out)
		assert.Equal(t, key, newKey)
	})

	t.Run("Empty output", func(t *testing.T) {
		newKey := f.parseKey("")
		assert.Equal(t, newKey, "")
	})
}
