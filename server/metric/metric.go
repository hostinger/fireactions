package metric

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hostinger/fireactions/server/store"
	"github.com/hostinger/fireactions/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var (
	namespace = "fireactions"
	subsystem = "server"
)

// PrometheusCollector is a Prometheus metric collector. It implements prometheus.Collector interface.
type PrometheusCollector struct {
	store store.Store

	runnersTotalMetric *prometheus.GaugeVec
	infoMetric         *prometheus.GaugeVec

	logger *zerolog.Logger
}

// PrometheusCollectorOpt is a functional option for the PrometheusCollector.
type PrometheusCollectorOpt func(*PrometheusCollector)

func WithLogger(logger *zerolog.Logger) PrometheusCollectorOpt {
	f := func(c *PrometheusCollector) {
		c.logger = logger
	}

	return f
}

// NewPrometheusCollector creates a new Prometheus metric collector which implements prometheus.Collector interface.
func NewPrometheusCollector(store store.Store, opts ...PrometheusCollectorOpt) *PrometheusCollector {
	runnersTotalMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Help:      "Total number of GitHub runners by state.",
		Name:      "runners_total",
	}, []string{"state"})

	infoMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Help:      "Information about the Fireactions server.",
		Name:      "info",
	}, []string{"version"})
	infoMetric.WithLabelValues(version.Version).Set(1)

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	c := &PrometheusCollector{
		store:              store,
		runnersTotalMetric: runnersTotalMetric,
		infoMetric:         infoMetric,
		logger:             &logger,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Describe implements prometheus.Collector interface.
func (c *PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	c.infoMetric.Describe(ch)
	c.runnersTotalMetric.Describe(ch)
}

// Collect implements prometheus.Collector interface.
func (c *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.updateMetrics(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msgf("failed to update metrics")
		return
	}

	c.infoMetric.Collect(ch)
	c.runnersTotalMetric.Collect(ch)
}

func (c *PrometheusCollector) updateRunnersTotalMetric(ctx context.Context) error {
	runners, err := c.store.GetRunners(ctx, nil)
	if err != nil {
		return err
	}

	c.runnersTotalMetric.
		With(prometheus.Labels{"state": "Pending"}).
		Set(0)
	c.runnersTotalMetric.
		With(prometheus.Labels{"state": "Idle"}).
		Set(0)
	c.runnersTotalMetric.
		With(prometheus.Labels{"state": "Active"}).
		Set(0)
	c.runnersTotalMetric.
		With(prometheus.Labels{"state": "Completed"}).
		Set(0)

	for _, runner := range runners {
		c.runnersTotalMetric.With(prometheus.Labels{"state": runner.Status.State.String()}).Inc()
	}

	return nil
}

func (c *PrometheusCollector) updateMetrics(ctx context.Context) error {
	err := c.updateRunnersTotalMetric(ctx)
	if err != nil {
		return fmt.Errorf("runners_total: %w", err)
	}

	return nil
}
