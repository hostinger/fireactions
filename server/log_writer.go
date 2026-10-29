package server

import "io"

// cappedWriter writes to w until limit bytes were written and discards the rest.
type cappedWriter struct {
	w       io.Writer
	limit   int64
	written int64
	failed  bool
}

func newCappedWriter(w io.Writer, limit int64) *cappedWriter {
	return &cappedWriter{w: w, limit: limit}
}

// Write never returns an error: the writer drains a pipe read by exec, and a
// failing writer would stop the reader and block the Firecracker process.
func (c *cappedWriter) Write(p []byte) (int, error) {
	if c.failed || c.written >= c.limit {
		return len(p), nil
	}

	chunk := p
	if remaining := c.limit - c.written; int64(len(chunk)) > remaining {
		chunk = chunk[:remaining]
	}

	n, err := c.w.Write(chunk)
	c.written += int64(n)
	if err != nil {
		c.failed = true
	}

	return len(p), nil
}
