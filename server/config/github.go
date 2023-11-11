package config

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/go-multierror"
)

// GitHubConfig is the configuration for the GitHub integration.
type GitHubConfig struct {
	JobLabelPrefix string                  `mapstructure:"job_label_prefix"`
	JobLabels      []*GitHubJobLabelConfig `mapstructure:"job_labels"`
	WebhookSecret  string                  `mapstructure:"webhook_secret"`
	AppID          int64                   `mapstructure:"app_id"`
	AppPrivateKey  string                  `mapstructure:"app_private_key"`
}

// GitHubJobLabelConfig is the configuration for a single job label. The label defines which repositories are allowed
// to use the label and how the jobs are executed.
type GitHubJobLabelConfig struct {
	Name                string   `mapstructure:"name"`
	AllowedRepositories []string `mapstructure:"allowed_repositories"`
	Runner              *RunnerConfig
}

// Validate validates the configuration.
func (c *GitHubConfig) Validate() error {
	var errs error

	if c.JobLabelPrefix == "" {
		errs = multierror.Append(errs, fmt.Errorf("job_label_prefix is required"))
	}

	if len(c.JobLabels) == 0 {
		errs = multierror.Append(errs, fmt.Errorf("at least one job_label is required"))
	}

	for _, label := range c.JobLabels {
		if err := label.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid job_label config: %w", err))
		}
	}

	if c.WebhookSecret == "" {
		errs = multierror.Append(errs, fmt.Errorf("webhook_secret is required"))
	}

	if c.AppID == 0 {
		errs = multierror.Append(errs, fmt.Errorf("app_id is required"))
	}

	if c.AppPrivateKey == "" {
		errs = multierror.Append(errs, fmt.Errorf("app_private_key is required"))
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
			if _, err := regexp.Compile(repo); err != nil {
				errs = multierror.Append(errs, fmt.Errorf("allowed_repositories regexp is invalid: %w", err))
			}
		}
	}

	if c.Runner == nil {
		errs = multierror.Append(errs, fmt.Errorf("runner config is required"))
	} else {
		if err := c.Runner.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid runner config: %w", err))
		}
	}

	return errs
}
