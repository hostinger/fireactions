package config

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// Config is the configuration for the Server.
type Config struct {
	HTTP         *HTTPConfig   `mapstructure:"http"`
	DataDir      string        `mapstructure:"data-dir"`
	GitHubConfig *GitHubConfig `mapstructure:"github"`
	LogLevel     string        `mapstructure:"log-level"`
}

// NewDefaultConfig creates a new default Config.
func NewDefaultConfig() *Config {
	cfg := &Config{
		HTTP:         &HTTPConfig{ListenAddress: ":8080"},
		DataDir:      "/var/lib/fireactions",
		GitHubConfig: &GitHubConfig{JobLabelPrefix: "fireactions"},
		LogLevel:     "info",
	}

	return cfg
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	var errs error

	if c.HTTP == nil {
		errs = multierror.Append(errs, fmt.Errorf("http is required"))
	} else {
		if err := c.HTTP.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid http config: %w", err))
		}
	}

	if c.DataDir == "" {
		errs = multierror.Append(errs, fmt.Errorf("data-dir is required"))
	}

	if c.GitHubConfig == nil {
		errs = multierror.Append(errs, fmt.Errorf("github is required"))
	} else {
		if err := c.GitHubConfig.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid github config: %w", err))
		}
	}

	switch c.LogLevel {
	case "trace", "debug", "info", "warn", "error":
	case "":
		errs = multierror.Append(errs, fmt.Errorf("log-level is required"))
	default:
		errs = multierror.Append(errs, fmt.Errorf("log-level must be one of: trace, debug, info, warn, error"))
	}

	return errs
}
