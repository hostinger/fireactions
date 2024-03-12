package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	config, err := NewConfig("testdata/config1.yaml")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	assert.Equal(t, "testdata/config1.yaml", config.path)
}
