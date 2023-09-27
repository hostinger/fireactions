package server

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/internal/server/scheduler"
)

// Config is the configuration for the Server.
type Config struct {
	ListenAddr string            `mapstructure:"listen-addr"`
	GitHub     *GitHubConfig     `mapstructure:"github"`
	Scheduler  *scheduler.Config `mapstructure:"scheduler"`
	LogLevel   string            `mapstructure:"log-level"`
	DataDir    string            `mapstructure:"data-dir"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	var err error

	if c.ListenAddr == "" {
		err = multierror.Append(err, errors.New("Config.ListenAddr is required, but was not provided"))
	}

	if c.GitHub == nil {
		err = multierror.Append(err, errors.New("Config.GitHub is required, but was not provided"))
	}

	if err := c.GitHub.Validate(); err != nil {
		err = multierror.Append(err, fmt.Errorf("Config.GitHub is invalid: %w", err))
	}

	if c.Scheduler == nil {
		err = multierror.Append(err, errors.New("Config.Scheduler is required, but was not provided"))
	}

	if err := c.Scheduler.Validate(); err != nil {
		err = multierror.Append(err, fmt.Errorf("Config.Scheduler is invalid: %w", err))
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error", "fatal", "panic":
	default:
		err = multierror.Append(err, fmt.Errorf("Config.LogLevel is invalid (%s). Must be one of: debug, info, warn, error, fatal, panic", c.LogLevel))
	}

	if c.DataDir == "" {
		err = multierror.Append(err, errors.New("Config.DataDir is required, but was not provided"))
	}

	return err
}

// GitHubConfig is the configuration for the GitHub integration.
type GitHubConfig struct {
	JobLabelPrefix string `mapstructure:"job-label-prefix"`
	AppID          int64  `mapstructure:"app-id"`
	AppPrivateKey  string `mapstructure:"app-private-key"`
	WebhookSecret  string `mapstructure:"webhook-secret"`
}

// Validate validates the configuration.
func (c *GitHubConfig) Validate() error {
	var err error

	if c.JobLabelPrefix == "" {
		err = multierror.Append(err, errors.New("GitHub.JobLabelPrefix is required, but was not provided"))
	}

	if c.AppID == 0 {
		err = multierror.Append(err, errors.New("GitHub.AppID is required, but was not provided"))
	}

	if c.AppPrivateKey == "" {
		err = multierror.Append(err, errors.New("GitHub.AppPrivateKey is required, but was not provided"))
	}

	if c.WebhookSecret == "" {
		err = multierror.Append(err, errors.New("GitHub.WebhookSecret is required, but was not provided"))
	}

	return err
}
