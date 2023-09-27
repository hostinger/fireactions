package client

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hostinger/fireactions/client/runtime"
)

// Config is the configuration for the Client.
type Config struct {
	FireactionsServerURL string            `mapstructure:"fireactions_server_url"`
	Name                 string            `mapstructure:"name"`
	Labels               map[string]string `mapstructure:"labels"`
	CpuOvercommitRatio   float64           `mapstructure:"cpu_overcommit_ratio"`
	RamOvercommitRatio   float64           `mapstructure:"ram_overcommit_ratio"`
	ReconcileInterval    time.Duration     `mapstructure:"reconcile_interval"`
	ReconcileConcurrency int               `mapstructure:"reconcile_concurrency"`
	RuntimeConfig        *runtime.Config   `mapstructure:"runtime"`
	LogLevel             string            `mapstructure:"log_level"`
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	cfg := &Config{
		FireactionsServerURL: "http://127.0.0.1:8080",
		Name:                 os.Getenv("HOSTNAME"),
		CpuOvercommitRatio:   1,
		RamOvercommitRatio:   1,
		ReconcileInterval:    5 * time.Second,
		ReconcileConcurrency: 100,
		Labels:               map[string]string{},
		RuntimeConfig:        runtime.NewConfig(),
		LogLevel:             "info",
	}

	return cfg
}

// Validate validates the Config.
func (c *Config) Validate() error {
	var errs error

	if c.FireactionsServerURL == "" {
		errs = multierror.Append(errs, errors.New("fireactions_server_url is required"))
	}

	if c.Name == "" {
		errs = multierror.Append(errs, errors.New("name is required"))
	}

	if c.CpuOvercommitRatio <= 0 {
		errs = multierror.Append(errs, fmt.Errorf("cpu_overcommit_ratio (%f) must be greater than 0", c.CpuOvercommitRatio))
	}

	if c.RamOvercommitRatio <= 0 {
		errs = multierror.Append(errs, fmt.Errorf("ram_overcommit_ratio (%f) must be greater than 0", c.RamOvercommitRatio))
	}

	if c.ReconcileInterval <= 0 {
		errs = multierror.Append(errs, fmt.Errorf("reconcile_interval (%s) must be greater than 0", c.ReconcileInterval))
	}

	if c.ReconcileConcurrency <= 0 {
		errs = multierror.Append(errs, fmt.Errorf("reconcile_concurrency (%d) must be greater than 0", c.ReconcileConcurrency))
	}

	if c.RuntimeConfig == nil {
		errs = multierror.Append(errs, errors.New("runtime is required"))
	} else {
		if err := c.RuntimeConfig.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("runtime: %w", err))
		}
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error", "fatal", "panic", "trace":
	default:
		errs = multierror.Append(errs, fmt.Errorf("log_level (%s) is invalid", c.LogLevel))
	}

	return errs
}
