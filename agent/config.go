package agent

import "github.com/go-playground/validator/v10"

type Config struct {
	Port            uint32 `validate:"required"`
	RunnerJITConfig string `validate:"required"`
	Hostname        string `validate:"required"`
	LogLevel        string `validate:"required,oneof=debug info warn error fatal panic trace"`
	ShutdownOnExit  bool   `validate:""`
}

func (c Config) Validate() error {
	return validator.New().Struct(c)
}
