package agent

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/agent/runner"
	"github.com/rs/zerolog"
)

// ListenAddr is the address on which the Agent listens.
const ListenAddr = "0.0.0.0:6969"

// Agent represents a virtual machine agent that's responsible for running
// the actual GitHub runner.
type Agent struct {
	runner *runner.Runner
	config *Config
	server *http.Server
	stopCh chan struct{}
	logger *zerolog.Logger
}

// New creates a new Agent.
func New(config *Config) (*Agent, error) {
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	logLevel, _ := zerolog.ParseLevel(config.LogLevel)
	logger := zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().CallerWithSkipFrameCount(2).Logger()

	router := gin.New()
	server := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      router,
	}

	agent := &Agent{
		config: config,
		runner: nil,
		server: server,
		stopCh: make(chan struct{}),
		logger: &logger,
	}

	router.GET("/healthz", agent.healthzHandlerFunc())
	router.POST("/api/v1/start", agent.startHandlerFunc())
	router.POST("/api/v1/stop", agent.stopHandlerFunc())

	return agent, nil
}

// Start starts the Agent. It blocks until the Agent is stopped via Stop().
func (a *Agent) Start() error {
	a.logger.Info().Msgf("Starting agent on %s", ListenAddr)

	listener, err := net.Listen("tcp", ListenAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	err = a.server.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Stop stops the Agent.
func (a *Agent) Stop(ctx context.Context) error {
	a.logger.Info().Msg("Stopping agent...")
	return a.server.Shutdown(ctx)
}

// StartGitHubRunner starts a GitHub runner. If the GitHub runner is already running,
// it returns ErrAlreadyRunning.
func (a *Agent) StartGitHubRunner(ctx context.Context, name, url, token string, labels []string, opts ...runner.Opt) error {
	if a.runner != nil && a.runner.IsRunning() {
		return ErrAlreadyRunning
	}

	a.runner = runner.New(name, url, labels, opts...)

	err := a.runner.Configure(ctx, token)
	if err != nil {
		return fmt.Errorf("configure: %w", err)
	}

	return a.runner.Run(ctx)
}

// StopGitHubRunner stops a GitHub runner. If the GitHub runner is not running,
// it returns ErrNotRunning.
func (a *Agent) StopGitHubRunner(ctx context.Context, token string) error {
	if a.runner == nil || !a.runner.IsRunning() {
		return ErrNotRunning
	}

	err := a.runner.Stop(ctx)
	if err != nil {
		return err
	}

	return a.runner.Unconfigure(ctx, token)
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}
