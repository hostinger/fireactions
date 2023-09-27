package config

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// RunnerConfig is the configuration for the Runner.
type RunnerConfig struct {
	Image           string                  `mapstructure:"image"`
	ImagePullPolicy string                  `mapstructure:"image_pull_policy"`
	Affinity        []*RunnerAffinityConfig `mapstructure:"affinity"`
	Resources       *RunnerResourcesConfig  `mapstructure:"resources"`
}

// RunnerResourcesConfig is the configuration for the Runner resources.
type RunnerResourcesConfig struct {
	VCPUs    int64 `mapstructure:"vcpus"`
	MemoryMB int64 `mapstructure:"memory_mb"`
}

// RunnerAffinityConfig is the configuration for the Runner affinity.
type RunnerAffinityConfig struct {
	Key      string   `mapstructure:"key"`
	Operator string   `mapstructure:"operator"`
	Values   []string `mapstructure:"values"`
}

// Validate validates the configuration.
func (c *RunnerConfig) Validate() error {
	var errs error

	if c.Image == "" {
		errs = multierror.Append(errs, fmt.Errorf("image is required"))
	}

	switch c.ImagePullPolicy {
	case "Always", "IfNotPresent", "Never":
	default:
		errs = multierror.Append(errs, fmt.Errorf("image_pull_policy must be one of: Always, IfNotPresent, Never"))
	}

	for _, affinity := range c.Affinity {
		if err := affinity.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid affinity config: %w", err))
		}
	}

	if c.Resources == nil {
		errs = multierror.Append(errs, fmt.Errorf("resources config is required"))
	} else {
		if err := c.Resources.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid resources config: %w", err))
		}
	}

	return errs
}

// Validate validates the configuration.
func (c *RunnerAffinityConfig) Validate() error {
	var errs error

	if c.Key == "" {
		errs = multierror.Append(errs, fmt.Errorf("key is required"))
	}

	switch c.Operator {
	case "In", "NotIn":
	default:
		errs = multierror.Append(errs, fmt.Errorf("operator must be one of: In, NotIn"))
	}

	if len(c.Values) == 0 {
		errs = multierror.Append(errs, fmt.Errorf("values is required"))
	}

	for _, value := range c.Values {
		if value == "" {
			errs = multierror.Append(errs, fmt.Errorf("values must not contain empty strings"))
		}
	}

	return errs
}

// Validate validates the configuration.
func (c *RunnerResourcesConfig) Validate() error {
	var errs error

	if c.VCPUs <= 0 {
		errs = multierror.Append(errs, fmt.Errorf("vcpus must be greater than 0"))
	}

	if c.MemoryMB <= 0 {
		errs = multierror.Append(errs, fmt.Errorf("memory_mb must be greater than 0"))
	}

	if c.MemoryMB%256 != 0 {
		errs = multierror.Append(errs, fmt.Errorf("memory_mb must be a multiple of 256"))
	}

	return errs
}
