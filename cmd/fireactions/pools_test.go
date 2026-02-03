package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPoolsCommand_Structure(t *testing.T) {
	cmd := newPoolsCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "pools", cmd.Use)
	assert.Equal(t, "Manage pools", cmd.Short)

	subcommands := cmd.Commands()
	assert.NotEmpty(t, subcommands)

	subcommandNames := make([]string, 0)
	for _, subcmd := range subcommands {
		subcommandNames = append(subcommandNames, subcmd.Name())
	}

	assert.Contains(t, subcommandNames, "list")
	assert.Contains(t, subcommandNames, "pause")
	assert.Contains(t, subcommandNames, "resume")
	assert.Contains(t, subcommandNames, "scale")
}

func TestPoolsPauseCommand_Structure(t *testing.T) {
	cmd := newPoolsPauseCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "pause NAME", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestPoolsResumeCommand_Structure(t *testing.T) {
	cmd := newPoolsResumeCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "resume NAME", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestPoolsScaleCommand_Structure(t *testing.T) {
	cmd := newPoolsScaleCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "scale NAME --replicas N", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Flags().Lookup("replicas"))
}

func TestPoolsListCommand_Structure(t *testing.T) {
	cmd := newPoolsListCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "list", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}
