package imagegc

import (
	"errors"
	"time"

	"github.com/hashicorp/go-multierror"
)

// Config is the configuration for ImageGC.
type Config struct {
	Interval time.Duration
}

// NewDefaultConfig creates a new default Config.
func NewDefaultConfig() *Config {
	cfg := &Config{
		Interval: 5 * time.Minute,
	}

	return cfg
}

// Validate validates the configuration.
func (cfg *Config) Validate() error {
	var errs error
	if cfg.Interval <= 0 {
		errs = multierror.Append(errs, errors.New("interval must be greater than 0"))
	}

	return errs
}
