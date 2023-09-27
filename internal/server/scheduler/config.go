package scheduler

import (
	"errors"
)

// Config is the Scheduler configuration.
type Config struct {
	FreeCpuScorerMultiplier float64 `mapstructure:"free-cpu-scorer-multiplier"`
	FreeMemScorerMultiplier float64 `mapstructure:"free-mem-scorer-multiplier"`
}

// Validate validates the Scheduler configuration.
func (c *Config) Validate() error {
	if c.FreeCpuScorerMultiplier < 0 {
		return errors.New("free-cpu-scorer-multiplier must be >= 0")
	}

	if c.FreeMemScorerMultiplier < 0 {
		return errors.New("free-mem-scorer-multiplier must be >= 0")
	}

	return nil
}
