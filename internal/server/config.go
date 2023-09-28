package server

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/internal/server/scheduler"
	"github.com/hostinger/fireactions/internal/structs"
)

// Config is the configuration for the Server.
type Config struct {
	ListenAddr    string            `mapstructure:"listen-addr"`
	GitHub        *GitHubConfig     `mapstructure:"github"`
	Scheduler     *scheduler.Config `mapstructure:"scheduler"`
	LogLevel      string            `mapstructure:"log-level"`
	DefaultFlavor string            `mapstructure:"default-flavor"`
	Flavors       []*structs.Flavor `mapstructure:"flavors"`
	DefaultGroup  string            `mapstructure:"default-group"`
	Groups        []*structs.Group  `mapstructure:"groups"`
	DataDir       string            `mapstructure:"data-dir"`
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

	if c.DefaultFlavor == "" {
		err = multierror.Append(err, errors.New("Config.DefaultFlavor is required, but was not provided"))
	}

	if len(c.Flavors) == 0 {
		err = multierror.Append(err, errors.New("Config.Flavors is required, but was not provided"))
	}

	defaultFlavorExists := false
	for _, f := range c.Flavors {
		if f.Name == "" {
			err = multierror.Append(err, errors.New("Flavor.Name is required, but was not provided"))
		}

		if f.DiskSizeGB == 0 {
			err = multierror.Append(err, errors.New("Flavor.DiskSizeGB is required, but was not provided"))
		}

		if f.MemorySizeMB == 0 {
			err = multierror.Append(err, errors.New("Flavor.MemorySizeMB is required, but was not provided"))
		}

		if f.VCPUs == 0 {
			err = multierror.Append(err, errors.New("Flavor.VCPUs is required, but was not provided"))
		}

		if f.ImageName == "" {
			err = multierror.Append(err, errors.New("Flavor.Image is required, but was not provided"))
		}

		if f.Name == c.DefaultFlavor {
			defaultFlavorExists = true
		}
	}

	if !defaultFlavorExists {
		err = multierror.Append(err, fmt.Errorf("Config.DefaultFlavor (%s) does not exist in Config.Flavors", c.DefaultFlavor))
	}

	if c.DefaultGroup == "" {
		err = multierror.Append(err, errors.New("Config.DefaultGroup is required, but was not provided"))
	}

	if len(c.Groups) == 0 {
		err = multierror.Append(err, errors.New("Config.Groups is required, but was not provided"))
	}

	defaultGroupExists := false
	for _, g := range c.Groups {
		if strings.Contains(g.Name, "-") {
			err = multierror.Append(err, fmt.Errorf("Group.Name (%s) must not contain any hyphens", g.Name))
		}

		if g.Name == c.DefaultGroup {
			defaultGroupExists = true
		}
	}

	if !defaultGroupExists {
		err = multierror.Append(err, fmt.Errorf("Config.DefaultGroup (%s) does not exist in Config.Groups", c.DefaultGroup))
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
