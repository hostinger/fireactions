package main

import (
	"bytes"
	"testing"

	"github.com/hostinger/fireactions"
	"github.com/stretchr/testify/assert"
)

func TestVersionCmd(t *testing.T) {
	cmd := newVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, fireactions.Version)
	assert.Contains(t, output, "Built on")
	assert.Contains(t, output, "Git SHA")
}
