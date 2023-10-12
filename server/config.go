package server

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/server/scheduler"
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
	Images        []*ImageConfig    `mapstructure:"images"`
	DataDir       string            `mapstructure:"data-dir"`
}

// NewDefaultConfig creates a new default Config.
func NewDefaultConfig() *Config {
	cfg := &Config{
		ListenAddr: ":8080",
		GitHub: &GitHubConfig{
			JobLabelPrefix: "fireactions",
		},
		Scheduler: &scheduler.Config{
			FreeCpuScorerMultiplier: 1.0,
			FreeRamScorerMultiplier: 1.0,
		},
		DefaultFlavor: "default",
		Flavors: []*FlavorConfig{{
			Name:    "default",
			Image:   "ubuntu-22.04",
			Enabled: true,
			Disk:    50,
			Mem:     1024,
			CPU:     1,
		}},
		DefaultGroup: "default",
		Groups: []*GroupConfig{{
			Name:    "default",
			Enabled: true,
		}},
		DataDir:  "/var/lib/fireactions",
		LogLevel: "info",
	}

	return cfg
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	var errs error

	if c.ListenAddr == "" {
		errs = multierror.Append(errs, errors.New("listen-addr is required"))
	}

	if c.GitHub == nil {
		errs = multierror.Append(errs, errors.New("github config is required"))
	}

	if err := c.GitHub.Validate(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("github config is invalid: %w", err))
	}

	if err := c.Scheduler.Validate(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("scheduler config is invalid: %w", err))
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error", "fatal", "panic":
	default:
		errs = multierror.Append(errs, fmt.Errorf("log-level (%s) is invalid", c.LogLevel))
	}

	if c.DataDir == "" {
		errs = multierror.Append(errs, errors.New("data-dir is required"))
	}

	if c.DefaultFlavor == "" {
		errs = multierror.Append(errs, errors.New("default-flavor is required"))
	}

	if len(c.Flavors) == 0 {
		errs = multierror.Append(errs, errors.New("at least one flavor must be defined"))
	}

	defaultFlavorExists := false
	for _, f := range c.Flavors {
		err := f.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("flavor (%s) is invalid: %w", f.Name, err))
		}

		if f.Name == c.DefaultFlavor {
			defaultFlavorExists = true
		}
	}

	if !defaultFlavorExists {
		errs = multierror.Append(errs, fmt.Errorf("default-flavor (%s) does not exist in flavors", c.DefaultFlavor))
	}

	if c.DefaultGroup == "" {
		errs = multierror.Append(errs, errors.New("default-group is required"))
	}

	if len(c.Groups) == 0 {
		errs = multierror.Append(errs, errors.New("at least one group must be defined"))
	}

	defaultGroupExists := false
	for _, g := range c.Groups {
		err := g.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("group (%s) is invalid: %w", g.Name, err))
		}

		if g.Name == c.DefaultGroup {
			defaultGroupExists = true
		}
	}

	if !defaultGroupExists {
		errs = multierror.Append(errs, fmt.Errorf("default-group (%s) does not exist in groups", c.DefaultGroup))
	}

	if len(c.Images) == 0 {
		errs = multierror.Append(errs, errors.New("at least one image must be defined"))
	}

	for _, i := range c.Images {
		err := i.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("image (%s) is invalid: %w", i.Name, err))
		}
	}

	return errs
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

// FlavorConfig is the configuration for a Flavor.
type FlavorConfig struct {
	Name    string `mapstructure:"name"`
	Enabled bool   `mapstructure:"enabled"`
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

// GroupConfig is the configuration for a Group.
type GroupConfig struct {
	Name    string `mapstructure:"name"`
	Enabled bool   `mapstructure:"enabled"`
}

// Validate validates the configuration.
func (c *GroupConfig) Validate() error {
	var err error

	if c.Name == "" {
		err = multierror.Append(err, errors.New("name is required"))
	}

	return err
}

// ImageConfig is the configuration for an Image.
type ImageConfig struct {
	ID   string `mapstructure:"id"`
	Name string `mapstructure:"name"`
	URL  string `mapstructure:"url"`
}

// Validate validates the configuration.
func (c *ImageConfig) Validate() error {
	var err error

	if c.ID == "" {
		err = multierror.Append(err, errors.New("id is required"))
	}

	if c.Name == "" {
		err = multierror.Append(err, errors.New("name is required"))
	}

	if c.URL == "" {
		err = multierror.Append(err, errors.New("url is required"))
	}

	_, err = uuid.Parse(c.ID)
	if err != nil {
		err = multierror.Append(err, fmt.Errorf("id (%s) is invalid", c.ID))
	}

	return err
}
