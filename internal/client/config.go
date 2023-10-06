package client

import (
	"errors"
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
		err = multierror.Append(err, errors.New("server-url is required"))
	}

	if c.Organisation == "" {
		err = multierror.Append(err, errors.New("organisation is required"))
	}

	if c.Group == "" {
		err = multierror.Append(err, errors.New("group is required"))
	}

	if c.CpuOvercommitRatio < 1 {
		return errors.New("cpu-overcommit-ratio must be >= 1.0")
	}

	if c.MemOvercommitRatio < 1 {
		return errors.New("mem-overcommit-ratio must be >= 1.0")
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error", "fatal", "panic":
	default:
		err = multierror.Append(err, fmt.Errorf("log-level (%s) is invalid", c.LogLevel))
	}

	return err
}

func (c *Config) SetDefaults() {
	if c.CpuOvercommitRatio == 0 {
		c.CpuOvercommitRatio = 1.0
	}

	if c.MemOvercommitRatio == 0 {
		c.MemOvercommitRatio = 1.0
	}

	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
}
