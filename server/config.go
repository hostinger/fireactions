package server

import (
	"fmt"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// Config is the configuration for the Client.
type Config struct {
	BindAddress      string            `yaml:"bind_address" validate:"required,hostname_port"`
	Containerd       *ContainerdConfig `yaml:"containerd" validate:"required"`
	Metrics          *MetricsConfig    `yaml:"metrics"`
	BasicAuthEnabled bool              `yaml:"basic_auth_enabled" validate:""`
	BasicAuthUsers   map[string]string `yaml:"basic_auth_users" validate:"required_if=basic_auth_enabled true"`
	GitHub           *GitHubConfig     `yaml:"github" validate:"required"`
	Pools            []*PoolConfig     `yaml:"pools" validate:"required,min=1,dive,required"`
	LogLevel         string            `yaml:"log_level" validate:"required,oneof=debug info warn error fatal panic trace"`

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

type RunnerConfig struct {
	Name            string   `yaml:"name" validate:"required"`
	ImagePullPolicy string   `yaml:"image_pull_policy" validate:"required,oneof=Always Never IfNotPresent"`
	Image           string   `yaml:"image" validate:"required"`
	Organization    string   `yaml:"organization" validate:"required"`
	GroupID         int64    `yaml:"group_id" validate:"required"`
	Labels          []string `yaml:"labels" validate:"required"`
}

type FirecrackerConfig struct {
	BinaryPath       string                             `yaml:"binary_path" `
	KernelImagePath  string                             `yaml:"kernel_image_path"`
	KernelArgs       string                             `yaml:"kernel_args"`
	MachineConfig    FirecrackerMachineConfig           `yaml:"machine_config"`
	NetworkInterface *FirecrackerNetworkInterfaceConfig `yaml:"network_interface"`
	Metadata         map[string]interface{}             `yaml:"metadata"`
}

type FirecrackerMachineConfig struct {
	VcpuCount  int64 `yaml:"vcpu_count"`
	MemSizeMib int64 `yaml:"mem_size_mib"`
}

// FirecrackerNetworkInterfaceConfig configures the MicroVM's network interface.
// Rate limiters are optional, a nil limiter leaves that direction unlimited.
type FirecrackerNetworkInterfaceConfig struct {
	InRateLimiter  *FirecrackerRateLimiterConfig `yaml:"in_rate_limiter"`
	OutRateLimiter *FirecrackerRateLimiterConfig `yaml:"out_rate_limiter"`
}

// FirecrackerRateLimiterConfig defines an IO rate limiter with independent
// bytes/s and ops/s limits. A nil token bucket leaves that limit unlimited.
type FirecrackerRateLimiterConfig struct {
	Bandwidth *FirecrackerTokenBucketConfig `yaml:"bandwidth"`
	Ops       *FirecrackerTokenBucketConfig `yaml:"ops"`
}

// FirecrackerTokenBucketConfig defines a token bucket with a maximum capacity
// (Size), an optional initial burst size (OneTimeBurst) and the interval in
// milliseconds it takes to refill the bucket (RefillTime). The resulting rate
// is Size / RefillTime.
type FirecrackerTokenBucketConfig struct {
	Size         int64  `yaml:"size" validate:"required,gt=0"`
	OneTimeBurst *int64 `yaml:"one_time_burst" validate:"omitempty,gte=0"`
	RefillTime   int64  `yaml:"refill_time" validate:"required,gt=0"`
}

// toSDK converts the rate limiter configuration into its Firecracker SDK
// representation. Returns nil if the rate limiter isn't configured.
func (c *FirecrackerRateLimiterConfig) toSDK() *models.RateLimiter {
	if c == nil {
		return nil
	}

	return &models.RateLimiter{Bandwidth: c.Bandwidth.toSDK(), Ops: c.Ops.toSDK()}
}

// toSDK converts the token bucket configuration into its Firecracker SDK
// representation. Returns nil if the token bucket isn't configured.
func (c *FirecrackerTokenBucketConfig) toSDK() *models.TokenBucket {
	if c == nil {
		return nil
	}

	return &models.TokenBucket{
		Size:         firecracker.Int64(c.Size),
		RefillTime:   firecracker.Int64(c.RefillTime),
		OneTimeBurst: c.OneTimeBurst,
	}
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
	return validator.New().Struct(c)
}
