package client

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/client/imagegc"
	"github.com/hostinger/fireactions/client/imagesyncer"
)

// Config is the configuration for the Client.
type Config struct {
	ServerURL          string              `mapstructure:"server-url"`
	DataDir            string              `mapstructure:"data-dir"`
	Organisation       string              `mapstructure:"organisation"`
	Group              string              `mapstructure:"group"`
	CpuOvercommitRatio float64             `mapstructure:"cpu-overcommit-ratio"`
	MemOvercommitRatio float64             `mapstructure:"mem-overcommit-ratio"`
	ImageSyncer        *imagesyncer.Config `mapstructure:"image-syncer"`
	EnableImageGC      bool                `mapstructure:"enable-image-gc"`
	ImageGC            *imagegc.Config     `mapstructure:"image-gc"`
	LogLevel           string              `mapstructure:"log-level"`
}

// NewDefaultConfig creates a new default Config.
func NewDefaultConfig() *Config {
	cfg := &Config{
		CpuOvercommitRatio: 1.0,
		MemOvercommitRatio: 1.0,
		DataDir:            "/var/lib/fireactions",
		ImageSyncer:        imagesyncer.NewDefaultConfig(),
		ImageGC:            imagegc.NewDefaultConfig(),
		EnableImageGC:      true,
		LogLevel:           "info",
	}

	return cfg
}

// Validate validates the Config.
func (c *Config) Validate() error {
	var errs error

	if c.ServerURL == "" {
		errs = multierror.Append(errs, errors.New("server-url is required"))
	}

	if c.DataDir == "" {
		errs = multierror.Append(errs, errors.New("data-dir is required"))
	}

	if c.Organisation == "" {
		errs = multierror.Append(errs, errors.New("organisation is required"))
	}

	if c.Group == "" {
		errs = multierror.Append(errs, errors.New("group is required"))
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
		errs = multierror.Append(errs, fmt.Errorf("log-level (%s) is invalid", c.LogLevel))
	}

	err := c.ImageSyncer.Validate()
	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error validating image-syncer config: %w", err))
	}

	if c.EnableImageGC {
		err = c.ImageGC.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error validating image-gc config: %w", err))
		}
	}

	return errs
}
