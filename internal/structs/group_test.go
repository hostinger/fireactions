package structs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		g := &Group{Name: "test"}
		assert.Equal(t, "test", g.String())
	})

	t.Run("Equals", func(t *testing.T) {
		g1 := &Group{Name: "test"}
		g2 := &Group{Name: "test"}
		g3 := &Group{Name: "test2"}

		assert.True(t, g1.Equals(g2))
		assert.False(t, g1.Equals(g3))
	})
}
