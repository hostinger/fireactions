package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
)

// Config is the configuration for the Client.
type Config struct {
	FireactionsServerURL string             `mapstructure:"fireactions_server_url"`
	PollInterval         time.Duration      `mapstructure:"poll_interval"`
	Node                 *NodeConfig        `mapstructure:"node"`
	Firecracker          *FirecrackerConfig `mapstructure:"firecracker"`
	CNI                  *CNIConfig         `mapstructure:"cni"`
	Containerd           *ContainerdConfig  `mapstructure:"containerd"`
	LogLevel             string             `mapstructure:"log_level"`
}

// NewDefaultConfig creates a new default Config.
func NewDefaultConfig() *Config {
	cfg := &Config{
		FireactionsServerURL: "http://127.0.0.1:8080",
		PollInterval:         5 * time.Second,
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
		LogLevel: "info",
	}

	return cfg
}

// Validate validates the Config.
func (c *Config) Validate() error {
	var errs error

	if c.FireactionsServerURL == "" {
		errs = multierror.Append(errs, errors.New("fireactions_server_url is required"))
	}

	if c.PollInterval < 1*time.Second {
		errs = multierror.Append(errs, errors.New("poll_interval must be >= 1s"))
	}

	if c.Node == nil {
		errs = multierror.Append(errs, errors.New("node config is required"))
	} else {
		err := c.Node.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error validating node config: %w", err))
		}
	}

	if c.Firecracker == nil {
		errs = multierror.Append(errs, errors.New("firecracker config is required"))
	} else {
		err := c.Firecracker.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error validating firecracker config: %w", err))
		}
	}

	if c.CNI == nil {
		errs = multierror.Append(errs, errors.New("cni config is required"))
	} else {
		err := c.CNI.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error validating cni config: %w", err))
		}
	}

	if c.Containerd == nil {
		errs = multierror.Append(errs, errors.New("containerd config is required"))
	} else {
		err := c.Containerd.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error validating containerd config: %w", err))
		}
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error", "fatal", "panic":
	default:
		errs = multierror.Append(errs, fmt.Errorf("log_level (%s) is invalid", c.LogLevel))
	}

	return errs
}

type NodeConfig struct {
	Name               string            `mapstructure:"name"`
	CpuOvercommitRatio float64           `mapstructure:"cpu_overcommit_ratio"`
	RamOvercommitRatio float64           `mapstructure:"ram_overcommit_ratio"`
	Labels             map[string]string `mapstructure:"labels"`
}

func (c *NodeConfig) Validate() error {
	var errs error

	if c.CpuOvercommitRatio < 1 {
		errs = multierror.Append(errs, errors.New("cpu_overcommit_ratio must be >= 1.0"))
	}

	if c.RamOvercommitRatio < 1 {
		errs = multierror.Append(errs, errors.New("ram_overcommit_ratio must be >= 1.0"))
	}

	return errs
}

type FirecrackerConfig struct {
	BinaryPath      string `mapstructure:"binary_path"`
	KernelImagePath string `mapstructure:"kernel_image_path"`
	KernelArgs      string `mapstructure:"kernel_args"`
	SocketPath      string `mapstructure:"socket_path"`
	LogFilePath     string `mapstructure:"log_file_path"`
	LogLevel        string `mapstructure:"log_level"`
}

func (c *FirecrackerConfig) Validate() error {
	var errs error

	if c.BinaryPath == "" {
		errs = multierror.Append(errs, errors.New("binary_path is required"))
	}

	if c.KernelImagePath == "" {
		errs = multierror.Append(errs, errors.New("kernel_image_path is required"))
	}

	if c.KernelArgs == "" {
		errs = multierror.Append(errs, errors.New("kernel_args is required"))
	}

	if c.BinaryPath == "" {
		errs = multierror.Append(errs, errors.New("binary_path is required"))
	}

	if c.SocketPath == "" {
		errs = multierror.Append(errs, errors.New("socket_path is required"))
	}

	if c.LogFilePath == "" {
		errs = multierror.Append(errs, errors.New("log_file_path is required"))
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error", "fatal", "panic":
	default:
		errs = multierror.Append(errs, fmt.Errorf("log_level (%s) is invalid", c.LogLevel))
	}

	return errs
}

type CNIConfig struct {
	ConfDir string   `mapstructure:"conf_dir"`
	BinDirs []string `mapstructure:"bin_dirs"`
}

func (c *CNIConfig) Validate() error {
	var errs error

	if c.ConfDir == "" {
		errs = multierror.Append(errs, errors.New("conf_dir is required"))
	}

	if len(c.BinDirs) == 0 {
		errs = multierror.Append(errs, errors.New("at least one bin_dir is required"))
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
