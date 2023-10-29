package config

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// GitHubConfig is the configuration for the GitHub integration.
type GitHubConfig struct {
	DefaultJobLabel string                  `mapstructure:"default-job-label"`
	JobLabelPrefix  string                  `mapstructure:"job-label-prefix"`
	JobLabels       []*GitHubJobLabelConfig `mapstructure:"job-labels"`
	WebhookSecret   string                  `mapstructure:"webhook-secret"`
	AppID           int64                   `mapstructure:"app-id"`
	AppPrivateKey   string                  `mapstructure:"app-private-key"`
}

// GitHubJobLabelConfig is the configuration for a single job label. The label defines which repositories are allowed
// to use the label and how the jobs are executed.
type GitHubJobLabelConfig struct {
	Name                string   `mapstructure:"name"`
	AllowedRepositories []string `mapstructure:"allowed-repositories"`
	Runner              *RunnerConfig
}

// Validate validates the configuration.
func (c *GitHubConfig) Validate() error {
	var errs error

	if c.JobLabelPrefix == "" {
		errs = multierror.Append(errs, fmt.Errorf("job-label-prefix is required"))
	}

	if len(c.JobLabels) == 0 {
		errs = multierror.Append(errs, fmt.Errorf("at least one job-label is required"))
	}

	for _, label := range c.JobLabels {
		if err := label.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid job-label config: %w", err))
		}
	}

	if c.WebhookSecret == "" {
		errs = multierror.Append(errs, fmt.Errorf("webhook-secret is required"))
	}

	if c.AppID == 0 {
		errs = multierror.Append(errs, fmt.Errorf("app-id is required"))
	}

	if c.AppPrivateKey == "" {
		errs = multierror.Append(errs, fmt.Errorf("app-private-key is required"))
	}

	if c.DefaultJobLabel == "" {
		errs = multierror.Append(errs, fmt.Errorf("default-job-label is required"))
	} else {
		found := false
		for _, label := range c.JobLabels {
			if label.Name == c.DefaultJobLabel {
				found = true
				break
			}
		}

		if !found {
			errs = multierror.Append(errs, fmt.Errorf("default-job-label must be one of the configured job-labels"))
		}
	}

	return errs
}

// GetJobLabelConfig returns the job label config for the given label. If the label is not configured, the second
// return value is false.
func (c *GitHubConfig) GetJobLabelConfig(label string) (*GitHubJobLabelConfig, bool) {
	for _, c := range c.JobLabels {
		if c.Name != label {
			continue
		}

		return c, true
	}

	return nil, false
}

// Validate validates the configuration.
func (c *GitHubJobLabelConfig) Validate() error {
	var errs error

	if c.Name == "" {
		errs = multierror.Append(errs, fmt.Errorf("name is required"))
	}

	if len(c.AllowedRepositories) > 0 {
		for _, repo := range c.AllowedRepositories {
			if repo == "" {
				errs = multierror.Append(errs, fmt.Errorf("allowed-repositories must not contain empty strings"))
			}
		}
	}

	if c.Runner == nil {
		errs = multierror.Append(errs, fmt.Errorf("runner is required"))
	} else {
		if err := c.Runner.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid runner config: %w", err))
		}
	}

	return errs
}
