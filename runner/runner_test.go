package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	r := New("test")

	assert.Equal(t, "test", r.config)
	assert.Equal(t, defaultDir, r.directory)
	assert.NotNil(t, r.stdout)
	assert.NotNil(t, r.stderr)
	assert.NotNil(t, r.logger)
}

func TestWithStdout(t *testing.T) {
	r := New("test", WithStdout(nil))

	assert.Nil(t, r.stdout)
}

func TestWithStderr(t *testing.T) {
	r := New("test", WithStderr(nil))

	assert.Nil(t, r.stderr)
}

func TestWithLogger(t *testing.T) {
	r := New("test", WithLogger(nil))

	assert.Nil(t, r.logger)
}

func TestWithDirectory(t *testing.T) {
	r := New("test", WithDirectory("test"))

	assert.Equal(t, "test", r.directory)
}

func TestWithOwner(t *testing.T) {
	r := New("test", WithOwner("test"))

	assert.Equal(t, "test", r.owner)
}

func TestWithGroup(t *testing.T) {
	r := New("test", WithGroup("test"))

	assert.Equal(t, "test", r.group)
}
