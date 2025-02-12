package commands

import (
	"testing"

	"github.com/hostinger/fireactions"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cmd := New()

	assert.NotNil(t, cmd)
	assert.Equal(t, "fireactions", cmd.Use)
	assert.Equal(t, "BYOM (Bring Your Own Metal) and run self-hosted GitHub runners in ephemeral, fast and secure Firecracker based virtual machines.", cmd.Short)
	assert.True(t, cmd.SilenceErrors)
	assert.True(t, cmd.SilenceUsage)
	assert.Equal(t, fireactions.Version, cmd.Version)

	assert.NotNil(t, cmd.PersistentPreRun)
	assert.NotNil(t, cmd.FlagErrorFunc())
	assert.NotNil(t, cmd.VersionTemplate())
	assert.NotNil(t, cmd.CompletionOptions)
	assert.True(t, cmd.CompletionOptions.DisableDefaultCmd)

	assert.NotNil(t, cmd.PersistentFlags().Lookup("endpoint"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("username"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("password"))

	assert.NotNil(t, cmd.Commands())
	assert.Len(t, cmd.Commands(), 9) // 9 subcommands added
}
