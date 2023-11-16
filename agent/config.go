package agent

import "fmt"

// Config is the configuration for the Agent.
type Config struct {
	LogLevel string `mapstructure:"log_level"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.LogLevel == "" {
		return fmt.Errorf("log-level cannot be empty")
	}

	return nil
}
