package scheduler

import (
	"errors"
)

// Config is the Scheduler configuration.
type Config struct {
	FreeCpuScorerMultiplier float64 `mapstructure:"free-cpu-scorer-multiplier"`
	FreeRamScorerMultiplier float64 `mapstructure:"free-ram-scorer-multiplier"`
}

// Validate validates the Scheduler configuration.
func (c *Config) Validate() error {
	if c.FreeCpuScorerMultiplier < 0 {
		return errors.New("free-cpu-scorer-multiplier must be >= 0")
	}

	if c.FreeRamScorerMultiplier < 0 {
		return errors.New("free-mem-scorer-multiplier must be >= 0")
	}

	return nil
}

// SetDefaults sets the default values for the Scheduler configuration.
func (c *Config) SetDefaults() {
	if c.FreeCpuScorerMultiplier == 0 {
		c.FreeCpuScorerMultiplier = defaultFreeCpuScorerMultiplier
	}

	if c.FreeRamScorerMultiplier == 0 {
		c.FreeRamScorerMultiplier = defaultFreeRamScorerMultiplier
	}
}
