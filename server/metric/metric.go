package metric

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
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
	nodesTotalMetric   *prometheus.GaugeVec
	infoMetric         *prometheus.GaugeVec
	nodeInfoMetric     *prometheus.GaugeVec

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

	nodesTotalMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Help:      "Total number of nodes by state.",
		Name:      "nodes_total",
	}, []string{"state"})

	nodesInfoMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Help:      "Information about the Fireactions node",
		Name:      "node_info",
	}, []string{"id", "name", "labels", "state", "cpu_overcommit_ratio", "ram_overcommit_ratio", "reconcile_interval"})

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
		nodesTotalMetric:   nodesTotalMetric,
		nodeInfoMetric:     nodesInfoMetric,
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
	c.nodeInfoMetric.Describe(ch)
	c.nodesTotalMetric.Describe(ch)
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
	c.nodeInfoMetric.Collect(ch)
	c.nodesTotalMetric.Collect(ch)
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

func (c *PrometheusCollector) updateNodesTotalMetric(ctx context.Context) error {
	nodes, err := c.store.GetNodes(ctx, nil)
	if err != nil {
		return err
	}

	c.nodeInfoMetric.Reset()
	c.nodesTotalMetric.
		With(prometheus.Labels{"state": "Ready"}).
		Set(0)
	c.nodesTotalMetric.
		With(prometheus.Labels{"state": "NotReady"}).
		Set(0)

	for _, node := range nodes {
		c.nodesTotalMetric.With(prometheus.Labels{"state": node.Status.State.String()}).Inc()

		labels := make([]string, 0, len(node.Labels))
		for k, v := range node.Labels {
			labels = append(labels, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(labels)

		c.nodeInfoMetric.With(prometheus.Labels{
			"id":                   node.ID,
			"name":                 node.Name,
			"labels":               strings.Join(labels, ","),
			"state":                node.Status.State.String(),
			"cpu_overcommit_ratio": fmt.Sprintf("%v", node.CPU.OvercommitRatio),
			"ram_overcommit_ratio": fmt.Sprintf("%v", node.RAM.OvercommitRatio),
			"reconcile_interval":   fmt.Sprintf("%v", node.ReconcileInterval),
		}).Set(1)
	}

	return nil
}

func (c *PrometheusCollector) updateMetrics(ctx context.Context) error {
	err := c.updateRunnersTotalMetric(ctx)
	if err != nil {
		return fmt.Errorf("runners_total: %w", err)
	}

	err = c.updateNodesTotalMetric(ctx)
	if err != nil {
		return fmt.Errorf("nodes_total: %w", err)
	}

	return nil
}
