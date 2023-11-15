package config

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
)

// Config is the configuration for the Server.
type Config struct {
	HTTP         *HTTPConfig   `mapstructure:"http"`
	DataDir      string        `mapstructure:"data_dir"`
	GitHubConfig *GitHubConfig `mapstructure:"github"`
	LogLevel     string        `mapstructure:"log_level"`
}

// NewDefaultConfig creates a new default Config.
func NewDefaultConfig() *Config {
	cfg := &Config{
		HTTP:         &HTTPConfig{ListenAddress: ":8080"},
		DataDir:      "/var/lib/fireactions",
		GitHubConfig: &GitHubConfig{JobLabelPrefix: "fireactions-"},
		LogLevel:     "info",
	}

	return cfg
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	var errs error

	if c.HTTP == nil {
		errs = multierror.Append(errs, fmt.Errorf("http config is required"))
	} else {
		if err := c.HTTP.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid http config: %w", err))
		}
	}

	if c.DataDir == "" {
		errs = multierror.Append(errs, fmt.Errorf("data_dir is required"))
	} else {
		_, err := os.Stat(c.DataDir)
		if err != nil && os.IsNotExist(err) {
			errs = multierror.Append(errs, fmt.Errorf("data_dir does not exist: %s", c.DataDir))
		}
	}

	if c.GitHubConfig == nil {
		errs = multierror.Append(errs, fmt.Errorf("github config is required"))
	} else {
		if err := c.GitHubConfig.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid github config: %w", err))
		}
	}

	switch c.LogLevel {
	case "trace", "debug", "info", "warn", "error":
	case "":
		errs = multierror.Append(errs, fmt.Errorf("log_level is required"))
	default:
		errs = multierror.Append(errs, fmt.Errorf("log_level must be one of: trace, debug, info, warn, error"))
	}

	return errs
}
