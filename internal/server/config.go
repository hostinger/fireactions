package server

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/internal/server/scheduler"
)

// Config is the configuration for the Server.
type Config struct {
	ListenAddr    string            `mapstructure:"listen-addr"`
	GitHub        *GitHubConfig     `mapstructure:"github"`
	Scheduler     *scheduler.Config `mapstructure:"scheduler"`
	LogLevel      string            `mapstructure:"log-level"`
	DefaultFlavor string            `mapstructure:"default-flavor"`
	Flavors       []*FlavorConfig   `mapstructure:"flavors"`
	DefaultGroup  string            `mapstructure:"default-group"`
	Groups        []*GroupConfig    `mapstructure:"groups"`
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
		err := f.Validate()
		if err != nil {
			err = multierror.Append(err, fmt.Errorf("Config.Flavors is invalid: %w", err))
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
		err := g.Validate()
		if err != nil {
			err = multierror.Append(err, fmt.Errorf("Config.Groups is invalid: %w", err))
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

// SetDefaults sets the default values for the configuration.
func (c *Config) SetDefaults() {
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}

	if c.DataDir == "" {
		c.DataDir = "/var/lib/fireactions"
	}

	c.GitHub.SetDefaults()
	for _, f := range c.Flavors {
		f.SetDefaults()
	}
	for _, g := range c.Groups {
		g.SetDefaults()
	}

	c.Scheduler.SetDefaults()
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

// SetDefaults sets the default values for the configuration.
func (c *GitHubConfig) SetDefaults() {
	if c.JobLabelPrefix == "" {
		c.JobLabelPrefix = "fireactions"
	}
}

// FlavorConfig is the configuration for a Flavor.
type FlavorConfig struct {
	Name    string `mapstructure:"name"`
	Enabled *bool  `mapstructure:"enabled"`
	Disk    int64  `mapstructure:"disk"`
	Mem     int64  `mapstructure:"mem"`
	CPU     int64  `mapstructure:"cpu"`
	Image   string `mapstructure:"image"`
}

// Validate validates the configuration.
func (c *FlavorConfig) Validate() error {
	var err error

	if c.Name == "" {
		err = multierror.Append(err, errors.New("Flavor.Name is required, but was not provided"))
	}

	if c.Disk == 0 {
		err = multierror.Append(err, errors.New("Flavor.Disk is required, but was not provided"))
	}

	if c.Mem == 0 {
		err = multierror.Append(err, errors.New("Flavor.Mem is required, but was not provided"))
	}

	if c.CPU == 0 {
		err = multierror.Append(err, errors.New("Flavor.CPU is required, but was not provided"))
	}

	if c.Image == "" {
		err = multierror.Append(err, errors.New("Flavor.Image is required, but was not provided"))
	}

	return err
}

// SetDefaults sets the default values for the configuration.
func (c *FlavorConfig) SetDefaults() {
	if c.Enabled == nil {
		b := true
		c.Enabled = &b
	}
}

// GroupConfig is the configuration for a Group.
type GroupConfig struct {
	Name    string `mapstructure:"name"`
	Enabled *bool  `mapstructure:"enabled"`
}

// Validate validates the configuration.
func (c *GroupConfig) Validate() error {
	var err error

	if c.Name == "" {
		err = multierror.Append(err, errors.New("Group.Name is required, but was not provided"))
	}

	if strings.Contains(c.Name, "-") {
		err = multierror.Append(err, fmt.Errorf("Group.Name (%s) must not contain any hyphens", c.Name))
	}

	return err
}

// SetDefaults sets the default values for the configuration.
func (c *GroupConfig) SetDefaults() {
	if c.Enabled == nil {
		b := true
		c.Enabled = &b
	}
}
