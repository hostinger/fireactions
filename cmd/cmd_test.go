package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cmd := New()

	assert.NotNil(t, cmd)
}
