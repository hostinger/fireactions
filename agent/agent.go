package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk/vsock"
	"github.com/hostinger/fireactions/agent/runner"
	agentv1 "github.com/hostinger/fireactions/proto/agent/v1"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
)

const (
	logFilePath = "/var/log/fireactions-agent.log"
)

type Agent struct {
	agentv1.UnimplementedAgentServiceServer
	cfg           Config
	logFile       string
	logFileWriter *os.File
	logger        *zerolog.Logger
	runner        *runner.Runner
}

type Opt func(a *Agent)

func New(cfg Config, opts ...Opt) (*Agent, error) {
	if cfgErr := cfg.Validate(); cfgErr != nil {
		return nil, fmt.Errorf("validate config: %w", cfgErr)
	}

	a := &Agent{
		cfg:     cfg,
		logFile: logFilePath,
	}

	for _, opt := range opts {
		opt(a)
	}

	if err := a.setupLogger(); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Agent) setupLogger() error {
	logDir := filepath.Dir(a.logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	logFileWriter, err := os.OpenFile(a.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	a.logFileWriter = logFileWriter

	logLevel, err := zerolog.ParseLevel(a.cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFileWriter)
	logger := zerolog.New(zerolog.ConsoleWriter{Out: multiWriter, TimeFormat: time.RFC3339}).With().
		Timestamp().
		Logger().Level(logLevel)

	a.logger = &logger
	return nil
}

// Close closes the agent resources, including the log file.
func (a *Agent) Close() error {
	if a.logFileWriter != nil {
		return a.logFileWriter.Close()
	}

	return nil
}

func (a *Agent) Run(ctx context.Context) error {
	if err := a.setHostname(); err != nil {
		return fmt.Errorf("setting hostname: %w", err)
	}

	// Run GitHub runner in background - it will trigger shutdown on success
	go a.runGitHubRunner(ctx)

	// Run gRPC server in main flow
	return a.runGRPCServer(ctx)
}

func (a *Agent) runGRPCServer(ctx context.Context) error {
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	logrusEntry := logrus.NewEntry(logrusLogger)

	listener, err := vsock.Listener(ctx, logrusEntry, a.cfg.Port)
	if err != nil {
		return fmt.Errorf("vsock listen: %w", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	agentv1.RegisterAgentServiceServer(grpcServer, a)

	errCh := make(chan error, 1)
	go func() {
		a.logger.Info().Msgf("Agent GRPC server listening on VSOCK port %d", a.cfg.Port)
		if err := grpcServer.Serve(listener); err != nil {
			errCh <- fmt.Errorf("grpc serve: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}

func (a *Agent) runGitHubRunner(ctx context.Context) {
	a.runner = runner.New(
		a.cfg.RunnerJITConfig,
		runner.WithLogger(a.logger),
	)

	if err := a.runner.Run(ctx); err != nil {
		a.logger.Error().Err(err).Msg("Runner encountered an error")
	}

	if !a.cfg.ShutdownOnExit {
		a.logger.Info().Msg("Runner completed, but shutdown on exit is disabled - keeping VM running")
		return
	}

	a.logger.Info().Msg("Runner completed, initiating VM shutdown")
	a.shutdown()
}

func (a *Agent) shutdown() {
	// We don't need to wait for it to complete since the VM will shut down anyway
	cmd := exec.Command("systemctl", "reboot")
	if err := cmd.Start(); err != nil {
		a.logger.Error().Err(err).Msg("Failed to initiate VM shutdown")
		return
	}

	a.logger.Info().Msg("Shutdown command executed")
}

func (a *Agent) setHostname() error {
	if err := unix.Sethostname([]byte(a.cfg.Hostname)); err != nil {
		return fmt.Errorf("sethostname: %w", err)
	}

	return nil
}
