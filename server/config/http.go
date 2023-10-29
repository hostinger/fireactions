package config

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// HTTPConfig is the configuration for the HTTP server.
type HTTPConfig struct {
	ListenAddress string `mapstructure:"listen-addr"`
}

// Validate validates the configuration.
func (c *HTTPConfig) Validate() error {
	var errs error

	if c.ListenAddress == "" {
		errs = multierror.Append(errs, fmt.Errorf("listen-addr is required"))
	}

	return errs
}
