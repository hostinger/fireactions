package server

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCappedWriter(t *testing.T) {
	var buf bytes.Buffer
	w := newCappedWriter(&buf, 10)

	n, err := w.Write([]byte("123456"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)

	n, err = w.Write([]byte("7890abc"))
	assert.NoError(t, err)
	assert.Equal(t, 7, n)

	n, err = w.Write([]byte("def"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)

	assert.Equal(t, "1234567890", buf.String())
}

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestCappedWriterSwallowsErrors(t *testing.T) {
	w := newCappedWriter(failingWriter{}, 10)

	n, err := w.Write([]byte("abc"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.True(t, w.failed)

	n, err = w.Write([]byte("def"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
}
