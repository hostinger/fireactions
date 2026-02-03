package main

import (
	"testing"

	"github.com/hostinger/fireactions"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "fireactions", cmd.Use)
	assert.Equal(t, "BYOM (Bring Your Own Metal) and run self-hosted GitHub runners in ephemeral, fast and secure Firecracker based virtual machines.", cmd.Short)
	assert.True(t, cmd.SilenceErrors)
	assert.True(t, cmd.SilenceUsage)
	assert.Equal(t, fireactions.Version, cmd.Version)

	assert.NotNil(t, cmd.FlagErrorFunc())
	assert.NotNil(t, cmd.VersionTemplate())

	assert.NotNil(t, cmd.Commands())
	assert.Len(t, cmd.Commands(), 9)
}
