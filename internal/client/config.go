package client

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

type Config struct {
	ServerURL          string  `mapstructure:"server-url"`
	Organisation       string  `mapstructure:"organisation"`
	Group              string  `mapstructure:"group"`
	CpuOvercommitRatio float64 `mapstructure:"cpu-overcommit-ratio"`
	MemOvercommitRatio float64 `mapstructure:"mem-overcommit-ratio"`
	LogLevel           string  `mapstructure:"log-level"`
}

func (c *Config) Validate() error {
	var err error

	if c.ServerURL == "" {
		err = multierror.Append(err, fmt.Errorf("Config.ServerURL is required, but was not provided"))
	}

	if c.Organisation == "" {
		err = multierror.Append(err, fmt.Errorf("Config.Organisation is required, but was not provided"))
	}

	if c.Group == "" {
		err = multierror.Append(err, fmt.Errorf("Config.Group is required, but was not provided"))
	}

	if c.CpuOvercommitRatio < 1 {
		return fmt.Errorf("Config.CpuOvercommitRatio must be >= 1.0")
	}

	if c.MemOvercommitRatio < 1 {
		return fmt.Errorf("Config.MemOvercommitRatio must be >= 1.0")
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error", "fatal", "panic":
	default:
		err = multierror.Append(err, fmt.Errorf("Config.LogLevel is invalid (%s). Must be one of: debug, info, warn, error, fatal, panic", c.LogLevel))
	}

	return err
}
