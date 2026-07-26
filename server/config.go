package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// defaultRunnerGroupID is the ID of the default GitHub runner group. Personal
// (user) accounts don't support runner groups and always use this ID.
const defaultRunnerGroupID = int64(1)

// Config is the configuration for the Client.
type Config struct {
	BindAddress      string            `yaml:"bind_address" validate:"required,hostname_port"`
	Containerd       *ContainerdConfig `yaml:"containerd" validate:"required"`
	Metrics          *MetricsConfig    `yaml:"metrics"`
	BasicAuthEnabled bool              `yaml:"basic_auth_enabled" validate:""`
	BasicAuthUsers   map[string]string `yaml:"basic_auth_users" validate:"required_if=basic_auth_enabled true"`
	GitHub           *GitHubConfig     `yaml:"github" validate:"required"`
	// "dive" is required. Without it the validator applies the tags to the
	// slice itself and never descends into the elements, so nothing inside a
	// pool - including the runner scope rules in RunnerConfig - is validated.
	Pools    []*PoolConfig `yaml:"pools" validate:"required,min=1,dive"`
	LogLevel string        `yaml:"log_level" validate:"required,oneof=debug info warn error fatal panic trace"`

	path string
}

type ContainerdConfig struct {
	Address   string `yaml:"address" validate:"required"`
	Namespace string `yaml:"namespace" validate:"required"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" validate:""`
	Address string `yaml:"address" validate:"required_if=enabled true,hostname_port"`
}

type GitHubConfig struct {
	AppPrivateKey string `yaml:"app_private_key" validate:"required"`
	AppID         int64  `yaml:"app_id" validate:"required"`
}

// RunnerConfig configures the GitHub runners of a Pool.
//
// Runners are registered either with an organization (Organization) or with a
// single repository (Repository). Exactly one of the two must be set.
// Repository scoped runners are the only option for personal (user) accounts,
// which can't have organization runners or runner groups.
type RunnerConfig struct {
	Name            string   `yaml:"name" validate:"required"`
	ImagePullPolicy string   `yaml:"image_pull_policy" validate:"required,oneof=Always Never IfNotPresent"`
	Image           string   `yaml:"image" validate:"required"`
	Organization    string   `yaml:"organization" validate:"required_without=Repository,excluded_with=Repository"`
	Repository      string   `yaml:"repository" validate:"required_without=Organization,omitempty,github_repository"`
	InstallationID  int64    `yaml:"installation_id"`
	GroupID         int64    `yaml:"group_id" validate:"required_without=Repository"`
	Labels          []string `yaml:"labels" validate:"required"`
}

// IsRepositoryScoped reports whether the runners are registered with a single
// repository instead of an organization.
func (c *RunnerConfig) IsRepositoryScoped() bool {
	return c.Repository != ""
}

// RepositoryOwnerAndName splits Repository into its owner and name parts. Both
// are empty if the runners aren't repository scoped.
func (c *RunnerConfig) RepositoryOwnerAndName() (string, string) {
	owner, name, _ := strings.Cut(c.Repository, "/")
	return owner, name
}

// Owner returns the GitHub account (organization or user) owning the runners.
func (c *RunnerConfig) Owner() string {
	if c.IsRepositoryScoped() {
		owner, _ := c.RepositoryOwnerAndName()
		return owner
	}

	return c.Organization
}

// Scope returns a human readable identifier of where the runners are
// registered, e.g. "hostinger" or "hostinger/fireactions".
func (c *RunnerConfig) Scope() string {
	if c.IsRepositoryScoped() {
		return c.Repository
	}

	return c.Organization
}

// RunnerGroupID returns the ID of the GitHub runner group to register runners
// in, falling back to the default group when none is configured.
func (c *RunnerConfig) RunnerGroupID() int64 {
	if c.GroupID == 0 {
		return defaultRunnerGroupID
	}

	return c.GroupID
}

type FirecrackerConfig struct {
	BinaryPath      string                   `yaml:"binary_path" `
	KernelImagePath string                   `yaml:"kernel_image_path"`
	KernelArgs      string                   `yaml:"kernel_args"`
	MachineConfig   FirecrackerMachineConfig `yaml:"machine_config"`
	Metadata        map[string]interface{}   `yaml:"metadata"`
}

type FirecrackerMachineConfig struct {
	VcpuCount  int64 `yaml:"vcpu_count"`
	MemSizeMib int64 `yaml:"mem_size_mib"`
}

// DefaultConfig creates a new Config with default values.
func DefaultConfig() *Config {
	c := &Config{
		BindAddress:      ":8080",
		Containerd:       &ContainerdConfig{Address: "/run/containerd/containerd.sock", Namespace: "fireactions"},
		Metrics:          &MetricsConfig{Enabled: true, Address: ":8081"},
		BasicAuthEnabled: false,
		BasicAuthUsers:   map[string]string{},
		GitHub:           &GitHubConfig{AppPrivateKey: "", AppID: 0},
		Pools:            []*PoolConfig{},
		LogLevel:         "debug",
	}

	return c
}

// NewConfigFromFile creates a new Config from a file.
func NewConfig(path string) (*Config, error) {
	c := DefaultConfig()
	c.path = path

	err := c.Load()
	if err != nil {
		return nil, err
	}

	err = c.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return c, nil
}

// LoadFromFile loads the configuration from a file.
func (c *Config) Load() error {
	file, err := os.OpenFile(c.path, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	return yaml.NewDecoder(file).Decode(c)
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	v := validator.New()

	err := v.RegisterValidation("github_repository", validateGitHubRepository)
	if err != nil {
		return fmt.Errorf("register validation: %w", err)
	}

	return v.Struct(c)
}

// validateGitHubRepository validates that a field holds a GitHub repository in
// <owner>/<repository> format.
func validateGitHubRepository(fl validator.FieldLevel) bool {
	owner, name, found := strings.Cut(fl.Field().String(), "/")
	return found && owner != "" && name != "" && !strings.Contains(name, "/")
}
