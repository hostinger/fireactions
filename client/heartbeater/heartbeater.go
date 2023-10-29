package heartbeater

import (
	"errors"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
)

// HeartbeatFunc is the function that will be called to send a heartbeat to the server.
type HeartbeatFunc func() error

// Config is the configuration for the Heartbeater.
type Config struct {
	// Interval is the interval at which the Heartbeater will send heartbeats to the server.
	Interval time.Duration

	// FailureThreshold is the number of consecutive failures after which the Heartbeater will
	// consider itself disconnected from the server.
	FailureThreshold int

	// SuccessThreshold is the number of consecutive successes after which the Heartbeater will
	// consider itself connected to the server.
	SuccessThreshold int

	// HeartbeatFunc is the function that will be called to send a heartbeat to the server.
	HeartbeatFunc HeartbeatFunc

	// FailureCh is the channel that will receive a value on heartbeat failure.
	FailureCh chan struct{}

	// SuccessCh is the channel that will receive a value on heartbeat success.
	SuccessCh chan struct{}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	var errs error

	if c.Interval == 0 {
		errs = multierror.Append(errs, errors.New("interval must be greater than 0"))
	}

	if c.FailureThreshold == 0 {
		errs = multierror.Append(errs, errors.New("failure-threshold must be greater than 0"))
	}

	if c.SuccessThreshold == 0 {
		errs = multierror.Append(errs, errors.New("success-threshold must be greater than 0"))
	}

	if c.SuccessThreshold > c.FailureThreshold {
		errs = multierror.Append(errs, errors.New("success-threshold must be less than or equal to failure-threshold"))
	}

	if c.HeartbeatFunc == nil {
		errs = multierror.Append(errs, errors.New("heartbeat-func is required"))
	}

	return errs
}

// Heartbeater sends heartbeats to the server.
type Heartbeater struct {
	config             *Config
	stopCh             chan struct{}
	consecutiveFailure int
	consecutiveSuccess int
	initial            bool
	failureCh          chan struct{}
	successCh          chan struct{}
	logger             *zerolog.Logger
}

// New creates a new Heartbeater.
func New(logger *zerolog.Logger, cfg *Config) (*Heartbeater, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	if cfg.FailureCh == nil {
		cfg.FailureCh = make(chan struct{})
	}

	if cfg.SuccessCh == nil {
		cfg.SuccessCh = make(chan struct{})
	}

	return newHeartbeater(logger, cfg), nil
}

func newHeartbeater(logger *zerolog.Logger, cfg *Config) *Heartbeater {
	h := &Heartbeater{
		config:             cfg,
		stopCh:             make(chan struct{}),
		consecutiveFailure: 0,
		consecutiveSuccess: 0,
		failureCh:          cfg.FailureCh,
		successCh:          cfg.SuccessCh,
		initial:            true,
		logger:             logger,
	}

	return h
}

// Start starts the Heartbeater. It will run in the background until Stop() is called.
func (h *Heartbeater) Run() {
	t := time.NewTicker(h.config.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			h.doHeartbeat()
		case <-h.stopCh:
			t.Stop()
			return
		}
	}
}

// Stop stops the Heartbeater.
func (h *Heartbeater) Stop() {
	close(h.stopCh)
}

func (h *Heartbeater) doHeartbeat() {
	err := h.config.HeartbeatFunc()
	if err != nil {
		h.handleHeartbeatFailure(err)
		return
	}

	h.handleHeartbeatSuccess()
}

func (h *Heartbeater) handleHeartbeatFailure(err error) {
	h.logger.Err(err).Msgf("error sending heartbeat")
	h.consecutiveFailure++
	if h.consecutiveFailure == h.config.FailureThreshold {
		h.failureCh <- struct{}{}
	}

	h.consecutiveSuccess = 0
}

func (h *Heartbeater) handleHeartbeatSuccess() {
	h.consecutiveSuccess++
	if h.consecutiveSuccess == h.config.SuccessThreshold {
		h.successCh <- struct{}{}
	}

	h.consecutiveFailure = 0
}
