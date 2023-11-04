package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPConfig_Validate(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		cfg := &HTTPConfig{}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "listen_addr is required")
	})
}
