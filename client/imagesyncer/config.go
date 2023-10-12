package imagesyncer

import (
	"errors"
	"time"

	"github.com/hashicorp/go-multierror"
)

// Config is the configuration for ImageSyncer.
type Config struct {
	Interval      time.Duration `mapstructure:"interval"`
	Images        []string      `mapstructure:"images"`
	MaxConcurrent int           `mapstructure:"max-concurrent"`
}

// NewDefaultConfig creates a new default Config.
func NewDefaultConfig() *Config {
	cfg := &Config{
		Interval:      5 * time.Second,
		Images:        []string{},
		MaxConcurrent: 5,
	}

	return cfg
}

// Validate validates the Config.
func (c *Config) Validate() error {
	var errs error
	if c.Interval <= 0 {
		errs = multierror.Append(errs, errors.New("interval must be greater than 0"))
	}

	if c.MaxConcurrent <= 0 {
		errs = multierror.Append(errs, errors.New("max-concurrent must be greater than 0"))
	}

	return errs
}

// DeepCopy creates a deep copy of the Config.
func (c *Config) DeepCopy() *Config {
	cfg := &Config{
		Interval:      c.Interval,
		Images:        make([]string, len(c.Images)),
		MaxConcurrent: c.MaxConcurrent,
	}

	copy(cfg.Images, c.Images)

	return cfg
}
