package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
)

// Config is the configuration for the Client.
type Config struct {
	FireactionsServerURL      string             `mapstructure:"fireactions-server-url"`
	HeartbeatInterval         time.Duration      `mapstructure:"heartbeat-interval"`
	HeartbeatSuccessThreshold int                `mapstructure:"heartbeat-success-threshold"`
	HeartbeatFailureThreshold int                `mapstructure:"heartbeat-failure-threshold"`
	PollInterval              time.Duration      `mapstructure:"poll-interval"`
	Node                      *NodeConfig        `mapstructure:"node"`
	Firecracker               *FirecrackerConfig `mapstructure:"firecracker"`
	CNI                       *CNIConfig         `mapstructure:"cni"`
	Containerd                *ContainerdConfig  `mapstructure:"containerd"`
	Metrics                   *MetricsConfig     `mapstructure:"metrics"`
	LogLevel                  string             `mapstructure:"log-level"`
}

// NewDefaultConfig creates a new default Config.
func NewDefaultConfig() *Config {
	cfg := &Config{
		FireactionsServerURL:      "http://127.0.0.1:8080",
		PollInterval:              5 * time.Second,
		HeartbeatSuccessThreshold: 3,
		HeartbeatFailureThreshold: 3,
		HeartbeatInterval:         1 * time.Second,
		Node: &NodeConfig{
			Name:               "",
			CpuOvercommitRatio: 1.0,
			RamOvercommitRatio: 1.0,
			Labels:             map[string]string{},
		},
		Firecracker: &FirecrackerConfig{
			BinaryPath:      "./firecracker",
			KernelImagePath: "vmlinux.bin",
			KernelArgs:      "console=ttyS0 noapic reboot=k panic=1 pci=off nomodules rw",
			SocketPath:      "/var/run/fireactions/%s.sock",
			LogLevel:        "info",
			LogFilePath:     "/var/log/fireactions/%s.log",
		},
		CNI: &CNIConfig{
			ConfDir: "/etc/cni/net.d",
			BinDirs: []string{"/opt/cni/bin"},
		},
		Containerd: &ContainerdConfig{
			Address: "/run/containerd/containerd.sock",
		},
		Metrics:  &MetricsConfig{Enabled: true, ListenAddr: "127.0.0.1:8081"},
		LogLevel: "info",
	}

	return cfg
}

// Validate validates the Config.
func (c *Config) Validate() error {
	var errs error

	if c.FireactionsServerURL == "" {
		errs = multierror.Append(errs, errors.New("fireactions-server-url is required"))
	}

	if c.HeartbeatInterval < 1*time.Second {
		errs = multierror.Append(errs, errors.New("heartbeat-interval must be >= 1s"))
	}

	if c.HeartbeatSuccessThreshold < 1 {
		errs = multierror.Append(errs, errors.New("heartbeat-success-threshold must be >= 1"))
	}

	if c.HeartbeatFailureThreshold < 1 {
		errs = multierror.Append(errs, errors.New("heartbeat-failure-threshold must be >= 1"))
	}

	if c.PollInterval < 1*time.Second {
		errs = multierror.Append(errs, errors.New("poll-interval must be >= 1s"))
	}

	if c.Node == nil {
		errs = multierror.Append(errs, errors.New("node is required"))
	} else {
		err := c.Node.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error validating node config: %w", err))
		}
	}

	if c.Firecracker == nil {
		errs = multierror.Append(errs, errors.New("firecracker is required"))
	} else {
		err := c.Firecracker.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error validating firecracker config: %w", err))
		}
	}

	if c.CNI == nil {
		errs = multierror.Append(errs, errors.New("cni is required"))
	} else {
		err := c.CNI.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error validating cni config: %w", err))
		}
	}

	if c.Containerd == nil {
		errs = multierror.Append(errs, errors.New("containerd is required"))
	} else {
		err := c.Containerd.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error validating containerd config: %w", err))
		}
	}

	if c.Metrics == nil {
		errs = multierror.Append(errs, errors.New("metrics is required"))
	} else {
		err := c.Metrics.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error validating metrics config: %w", err))
		}
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error", "fatal", "panic":
	default:
		errs = multierror.Append(errs, fmt.Errorf("log-level (%s) is invalid", c.LogLevel))
	}

	return errs
}

type NodeConfig struct {
	Name               string            `mapstructure:"name"`
	CpuOvercommitRatio float64           `mapstructure:"cpu-overcommit-ratio"`
	RamOvercommitRatio float64           `mapstructure:"ram-overcommit-ratio"`
	Labels             map[string]string `mapstructure:"labels"`
}

func (c *NodeConfig) Validate() error {
	var errs error

	if c.CpuOvercommitRatio < 1 {
		errs = multierror.Append(errs, errors.New("cpu-overcommit-ratio must be >= 1.0"))
	}

	if c.RamOvercommitRatio < 1 {
		errs = multierror.Append(errs, errors.New("ram-overcommit-ratio must be >= 1.0"))
	}

	return errs
}

type FirecrackerConfig struct {
	BinaryPath      string `mapstructure:"binary-path"`
	KernelImagePath string `mapstructure:"kernel-image-path"`
	KernelArgs      string `mapstructure:"kernel-args"`
	SocketPath      string `mapstructure:"socket-path"`
	LogFilePath     string `mapstructure:"log-file-path"`
	LogLevel        string `mapstructure:"log-level"`
}

func (c *FirecrackerConfig) Validate() error {
	var errs error

	if c.BinaryPath == "" {
		errs = multierror.Append(errs, errors.New("binary-path is required"))
	}

	if c.KernelImagePath == "" {
		errs = multierror.Append(errs, errors.New("kernel-image-path is required"))
	}

	if c.KernelArgs == "" {
		errs = multierror.Append(errs, errors.New("kernel-args is required"))
	}

	if c.BinaryPath == "" {
		errs = multierror.Append(errs, errors.New("binary-path is required"))
	}

	if c.SocketPath == "" {
		errs = multierror.Append(errs, errors.New("socket-path is required"))
	}

	if c.LogFilePath == "" {
		errs = multierror.Append(errs, errors.New("log-file-path is required"))
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error", "fatal", "panic":
	default:
		errs = multierror.Append(errs, fmt.Errorf("log-level (%s) is invalid", c.LogLevel))
	}

	return errs
}

type CNIConfig struct {
	ConfDir string   `mapstructure:"conf-dir"`
	BinDirs []string `mapstructure:"bin-dirs"`
}

func (c *CNIConfig) Validate() error {
	var errs error

	if c.ConfDir == "" {
		errs = multierror.Append(errs, errors.New("conf-dir is required"))
	}

	if len(c.BinDirs) == 0 {
		errs = multierror.Append(errs, errors.New("at least one bin-dir is required"))
	}

	return errs
}

type ContainerdConfig struct {
	Address string `mapstructure:"address"`
}

func (c *ContainerdConfig) Validate() error {
	var errs error

	if c.Address == "" {
		errs = multierror.Append(errs, errors.New("address is required"))
	}

	return errs
}

type MetricsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	ListenAddr string `mapstructure:"listen-addr"`
}

func (c *MetricsConfig) Validate() error {
	var errs error

	if c.Enabled && c.ListenAddr == "" {
		errs = multierror.Append(errs, errors.New("listen-addr is required"))
	}

	return errs
}
