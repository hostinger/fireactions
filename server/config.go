package server

import (
	"fmt"
	"os"
	"regexp"

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

// RunnerConfig is the configuration for the Runner.
type RunnerConfig struct {
	Image           string                  `mapstructure:"image"`
	ImagePullPolicy string                  `mapstructure:"image_pull_policy"`
	Metadata        map[string]interface{}  `mapstructure:"metadata"`
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

	return errs
}

// HTTPConfig is the configuration for the HTTP server.
type HTTPConfig struct {
	ListenAddress string `mapstructure:"listen_addr"`
}

// Validate validates the configuration.
func (c *HTTPConfig) Validate() error {
	var errs error

	if c.ListenAddress == "" {
		errs = multierror.Append(errs, fmt.Errorf("listen_addr is required"))
	}

	return errs
}

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
