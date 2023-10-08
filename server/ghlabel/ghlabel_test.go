package ghlabel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabel(t *testing.T) {
	t.Run("fireactions.us-east-1.1vcpu-2gb", func(t *testing.T) {
		label, err := New("fireactions.us-east-1.1vcpu-2gb")

		assert.NoError(t, err)
		assert.Equal(t, "fireactions", label.Prefix)
		assert.Equal(t, "us-east-1", label.Group)
		assert.Equal(t, "1vcpu-2gb", label.Flavor)
		assert.Equal(t, "fireactions.us-east-1.1vcpu-2gb", label.String())
	})

	t.Run("fireactions.us-east-1", func(t *testing.T) {
		label, err := New("fireactions.us-east-1")

		assert.NoError(t, err)
		assert.Equal(t, "fireactions", label.Prefix)
		assert.Equal(t, "us-east-1", label.Group)
		assert.Equal(t, "", label.Flavor)
		assert.Equal(t, "fireactions.us-east-1", label.String())
	})

	t.Run("fireactions", func(t *testing.T) {
		label, err := New("fireactions")

		assert.NoError(t, err)
		assert.Equal(t, "fireactions", label.Prefix)
		assert.Equal(t, "", label.Group)
		assert.Equal(t, "", label.Flavor)
		assert.Equal(t, "fireactions", label.String())
		assert.True(t, label.FlavorIsEmpty())
		assert.True(t, label.GroupIsEmpty())
	})

	t.Run("invalid", func(t *testing.T) {
		label, err := New("")

		assert.Error(t, err)
		assert.Nil(t, label)
	})
}
