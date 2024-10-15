package server

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// Config is the configuration for the Client.
type Config struct {
	BindAddress      string            `yaml:"bind_address" validate:"required,hostname_port"`
	Metrics          *MetricsConfig    `yaml:"metrics"`
	BasicAuthEnabled bool              `yaml:"basic_auth_enabled" validate:""`
	BasicAuthUsers   map[string]string `yaml:"basic_auth_users" validate:"required_if=basic_auth_enabled true"`
	GitHub           *GitHubConfig     `yaml:"github" validate:"required"`
	Pools            []*PoolConfig     `yaml:"pools" validate:"required,min=1"`
	LogLevel         string            `yaml:"log_level" validate:"required,oneof=debug info warn error fatal panic trace"`
	Debug            bool              `yaml:"debug" validate:""`

	path string
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" validate:""`
	Address string `yaml:"address" validate:"required_if=enabled true,hostname_port"`
}

type GitHubConfig struct {
	AppPrivateKey string `yaml:"app_private_key" validate:"required"`
	AppID         int64  `yaml:"app_id" validate:"required"`
}

type RunnerConfig struct {
	Name            string   `yaml:"name" validate:"required"`
	ImagePullPolicy string   `yaml:"image_pull_policy" validate:"required,oneof=always never ifnotpresent"`
	Image           string   `yaml:"image" validate:"required"`
	Organization    string   `yaml:"organization" validate:"required"`
	GroupID         int64    `yaml:"group_id" validate:"required"`
	Labels          []string `yaml:"labels" validate:"required"`
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
		Metrics:          &MetricsConfig{Enabled: true, Address: ":8081"},
		BasicAuthEnabled: false,
		BasicAuthUsers:   map[string]string{},
		GitHub:           &GitHubConfig{AppPrivateKey: "", AppID: 0},
		Pools:            []*PoolConfig{},
		LogLevel:         "debug",
		Debug:            false,
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

	defer file.Close()

	return yaml.NewDecoder(file).Decode(c)
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	return validator.New().Struct(c)
}
