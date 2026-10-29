package server

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"
	"time"

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

func TestCappedWriterDrainsProcessOutput(t *testing.T) {
	var buf bytes.Buffer
	w := newCappedWriter(&buf, 4096)

	cmd := exec.Command("sh", "-c", "head -c 20000000 /dev/zero")
	cmd.Stdout = w
	cmd.Stderr = w

	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(30 * time.Second):
		t.Fatal("process blocked on output")
	}

	assert.Equal(t, 4096, buf.Len())
}
