package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCommand_Structure(t *testing.T) {
	cmd := newValidateCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "validate <config-file>", cmd.Use)
	assert.Equal(t, "Validates the server configuration file", cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestValidateCommand_ValidConfig(t *testing.T) {
	cmd := newValidateCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	configPath := filepath.Join("..", "..", "server", "testdata", "config1.yaml")
	cmd.SetArgs([]string{configPath})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "is valid")
}

func TestValidateCommand_InvalidConfig(t *testing.T) {
	// Create a temporary invalid config file
	tmpFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	invalidConfig := `
bind_address: 0.0.0.0:8080
log_level: invalid_level
pools: []
`
	_, err = tmpFile.WriteString(invalidConfig)
	require.NoError(t, err)
	tmpFile.Close()

	cmd := newValidateCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{tmpFile.Name()})

	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestValidateCommand_MissingFile(t *testing.T) {
	cmd := newValidateCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{"/nonexistent/config.yaml"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestValidateCommand_NoArgs(t *testing.T) {
	cmd := newValidateCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}
