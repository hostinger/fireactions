package scheduler

import (
	"errors"
)

// Config is the Scheduler configuration.
type Config struct {
	FreeCpuScorerMultiplier float64 `mapstructure:"free-cpu-scorer-multiplier"`
	FreeRamScorerMultiplier float64 `mapstructure:"free-ram-scorer-multiplier"`
}

// NewDefaultConfig returns a new default Scheduler configuration.
func NewDefaultConfig() *Config {
	cfg := &Config{
		FreeCpuScorerMultiplier: 1.0,
		FreeRamScorerMultiplier: 1.0,
	}

	return cfg
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
