package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPConfig_Validate(t *testing.T) {
	t.Run("ReturnsErrorIfListenAddressIsEmpty", func(t *testing.T) {
		cfg := &HTTPConfig{
			ListenAddress: "",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "listen_addr is required")
	})
}
