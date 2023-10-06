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
		err = multierror.Append(err, errors.New("listen-addr is required"))
	}

	if c.GitHub == nil {
		err = multierror.Append(err, errors.New("github config is required"))
	}

	if err := c.GitHub.Validate(); err != nil {
		err = multierror.Append(err, fmt.Errorf("github config is invalid: %w", err))
	}

	if err := c.Scheduler.Validate(); err != nil {
		err = multierror.Append(err, fmt.Errorf("scheduler config is invalid: %w", err))
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error", "fatal", "panic":
	default:
		err = multierror.Append(err, fmt.Errorf("log-level (%s) is invalid", c.LogLevel))
	}

	if c.DataDir == "" {
		err = multierror.Append(err, errors.New("data-dir is required"))
	}

	if c.DefaultFlavor == "" {
		err = multierror.Append(err, errors.New("default-flavor is required"))
	}

	if len(c.Flavors) == 0 {
		err = multierror.Append(err, errors.New("at least one flavor must be defined"))
	}

	defaultFlavorExists := false
	for _, f := range c.Flavors {
		err := f.Validate()
		if err != nil {
			err = multierror.Append(err, fmt.Errorf("flavor (%s) is invalid: %w", f.Name, err))
		}

		if f.Name == c.DefaultFlavor {
			defaultFlavorExists = true
		}
	}

	if !defaultFlavorExists {
		err = multierror.Append(err, fmt.Errorf("default-flavor (%s) does not exist in flavors", c.DefaultFlavor))
	}

	if c.DefaultGroup == "" {
		err = multierror.Append(err, errors.New("default-group is required"))
	}

	if len(c.Groups) == 0 {
		err = multierror.Append(err, errors.New("at least one group must be defined"))
	}

	defaultGroupExists := false
	for _, g := range c.Groups {
		err := g.Validate()
		if err != nil {
			err = multierror.Append(err, fmt.Errorf("group (%s) is invalid: %w", g.Name, err))
		}

		if g.Name == c.DefaultGroup {
			defaultGroupExists = true
		}
	}

	if !defaultGroupExists {
		err = multierror.Append(err, fmt.Errorf("default-group (%s) does not exist in groups", c.DefaultGroup))
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

	if c.Scheduler == nil {
		c.Scheduler = &scheduler.Config{}
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
		err = multierror.Append(err, errors.New("job-label-prefix is required"))
	}

	if c.AppID == 0 {
		err = multierror.Append(err, errors.New("app-id is required"))
	}

	if c.AppPrivateKey == "" {
		err = multierror.Append(err, errors.New("app-private-key is required"))
	}

	if c.WebhookSecret == "" {
		err = multierror.Append(err, errors.New("webhook-secret is required"))
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
		err = multierror.Append(err, errors.New("name is required"))
	}

	if c.Disk == 0 {
		err = multierror.Append(err, errors.New("disk is required"))
	}

	if c.Mem == 0 {
		err = multierror.Append(err, errors.New("mem is required"))
	}

	if c.CPU == 0 {
		err = multierror.Append(err, errors.New("cpu is required"))
	}

	if c.Image == "" {
		err = multierror.Append(err, errors.New("image is required"))
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
		err = multierror.Append(err, errors.New("name is required"))
	}

	if strings.Contains(c.Name, "-") {
		err = multierror.Append(err, errors.New("name cannot contain dashes"))
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
