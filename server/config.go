package server

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions"
	"github.com/samber/lo"
)

// Config is the configuration for the Server.
type Config struct {
	HTTP      *HTTPConfig       `mapstructure:"http"`
	DataDir   string            `mapstructure:"data_dir"`
	GitHub    *GitHubConfig     `mapstructure:"github"`
	Metrics   *MetricsConfig    `mapstructure:"metrics"`
	History   *HistoryConfig    `mapstructure:"history"`
	JobLabels []*JobLabelConfig `mapstructure:"job_labels"`
	LogLevel  string            `mapstructure:"log_level"`
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	cfg := &Config{
		HTTP:      &HTTPConfig{ListenAddress: ":8080"},
		DataDir:   "",
		GitHub:    &GitHubConfig{WebhookSecret: "", AppID: 0, AppPrivateKey: ""},
		Metrics:   &MetricsConfig{Enabled: true, Address: ":8081"},
		History:   &HistoryConfig{MaxWorkflowRunAge: time.Hour * 24 * 7},
		JobLabels: []*JobLabelConfig{},
		LogLevel:  "info",
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

	if c.GitHub == nil {
		errs = multierror.Append(errs, fmt.Errorf("github config is required"))
	} else {
		if err := c.GitHub.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid github config: %w", err))
		}
	}

	if c.Metrics == nil {
		errs = multierror.Append(errs, fmt.Errorf("metrics config is required"))
	} else {
		if err := c.Metrics.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid metrics config: %w", err))
		}
	}

	for _, label := range c.JobLabels {
		if err := label.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("invalid job_label (%s): %w", label.Name, err))
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

// FindJobLabel finds a JobLabel by name.
func (c *Config) FindJobLabel(name string) (*JobLabelConfig, bool) {
	for _, label := range c.JobLabels {
		if label.Name != name {
			continue
		}

		return label, true
	}

	return nil, false
}

func (c *Config) FindMatchingJobLabel(labels []string) (*JobLabelConfig, bool) {
	for _, label := range labels {
		jobLabel, ok := c.FindJobLabel(label)
		if !ok {
			continue
		}

		return jobLabel, true
	}

	return nil, false
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
	WebhookSecret string `mapstructure:"webhook_secret"`
	AppPrivateKey string `mapstructure:"app_private_key"`
	AppID         int64  `mapstructure:"app_id"`
}

// Validate validates the configuration.
func (c *GitHubConfig) Validate() error {
	var errs error

	if c.WebhookSecret == "" {
		errs = multierror.Append(errs, fmt.Errorf("webhook_secret is required"))
	}

	if c.AppPrivateKey == "" {
		errs = multierror.Append(errs, fmt.Errorf("app_private_key is required"))
	}

	if c.AppID == 0 {
		errs = multierror.Append(errs, fmt.Errorf("app_id is required"))
	}

	return errs
}

// JobLabelConfig is the configuration for a single job label. The label defines which repositories are allowed
// to use the label and how the jobs are executed.
type JobLabelConfig struct {
	Name                  string                            `mapstructure:"name"`
	AllowedRepositories   []string                          `mapstructure:"allowed_repositories"`
	RunnerLabels          []string                          `mapstructure:"runner_labels"`
	RunnerNameTemplate    string                            `mapstructure:"runner_name_template"`
	RunnerImage           string                            `mapstructure:"runner_image"`
	RunnerImagePullPolicy fireactions.RunnerImagePullPolicy `mapstructure:"runner_image_pull_policy"`
	RunnerResources       fireactions.RunnerResources       `mapstructure:"runner_resources"`
	RunnerAffinity        []*fireactions.RunnerAffinityRule `mapstructure:"runner_affinity"`
	RunnerMetadata        map[string]interface{}            `mapstructure:"runner_metadata"`
}

// Validate validates the configuration.
func (c *JobLabelConfig) Validate() error {
	var errs error

	if c.Name == "" {
		errs = multierror.Append(errs, fmt.Errorf("name is required"))
	}

	if len(c.AllowedRepositories) == 0 {
		errs = multierror.Append(errs, fmt.Errorf("allowed_repositories is required"))
	}

	if c.RunnerNameTemplate == "" {
		errs = multierror.Append(errs, fmt.Errorf("runner_name_template is required"))
	}

	if _, err := template.New("runner_name").Parse(c.RunnerNameTemplate); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("invalid runner_name_template: %w", err))
	}

	if c.RunnerImage == "" {
		errs = multierror.Append(errs, fmt.Errorf("runner_image is required"))
	}

	if c.RunnerImagePullPolicy == "" {
		errs = multierror.Append(errs, fmt.Errorf("runner_image_pull_policy is required"))
	}

	if err := c.RunnerResources.Validate(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("invalid runner_resources: %w", err))
	}

	return errs
}

// MustGetRunnerName renders the runner name from the template. It panics if the template is invalid.
func (c *JobLabelConfig) MustGetRunnerName(runnerID string) string {
	name, err := c.GetRunnerName(runnerID)
	if err != nil {
		panic(err)
	}

	return name
}

func (c *JobLabelConfig) IsAllowedRepository(repository string) bool {
	return lo.ContainsBy(c.AllowedRepositories, func(repository string) bool {
		regexp, err := regexp.Compile(repository)
		if err != nil {
			return false
		}

		return regexp.MatchString(repository)
	})
}

// GetRunnerName renders the runner name from the template.
func (c *JobLabelConfig) GetRunnerName(runnerID string) (string, error) {
	templ, err := template.New("runner_name").Parse(c.RunnerNameTemplate)
	if err != nil {
		return "", fmt.Errorf("invalid runner_name_template: %w", err)
	}

	var buf strings.Builder
	err = templ.Execute(&buf, map[string]interface{}{
		"ID": runnerID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute runner_name_template: %w", err)
	}

	return buf.String(), nil
}

// GetRunnerLabels returns the labels for the runner. It includes the fireactions label and the job label.
func (c *JobLabelConfig) GetRunnerLabels() []string {
	labels := []string{}
	labels = append(labels, c.Name)
	labels = append(labels, c.RunnerLabels...)

	return labels
}

// MetricsConfig is the configuration for the metrics server.
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Address string `mapstructure:"address"`
}

// Validate validates the configuration.
func (c *MetricsConfig) Validate() error {
	var errs error

	if c.Address == "" {
		errs = multierror.Append(errs, fmt.Errorf("address is required"))
	}

	return errs
}

// HistoryConfig is the configuration for the garbage collection various history data.
type HistoryConfig struct {
	MaxWorkflowRunAge time.Duration `mapstructure:"max_workflow_run_age"`
}

// Validate validates the configuration.
func (c *HistoryConfig) Validate() error {
	var errs error

	if c.MaxWorkflowRunAge == 0 {
		errs = multierror.Append(errs, fmt.Errorf("max_workflow_run_age is required"))
	}

	return errs
}
