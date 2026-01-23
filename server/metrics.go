package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "fireactions"
)

var (
	metricUp = promauto.NewGauge(prometheus.GaugeOpts{
		Name:      "server_up",
		Namespace: namespace,
		Help:      "Is the server up",
	})

	metricPoolsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name:      "pools_total",
		Namespace: namespace,
		Help:      "Total number of pools",
	})

	metricPoolRunnersCurrent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "pool_runners_current",
		Namespace: namespace,
		Help:      "Current number of running runners in a pool",
	}, []string{"pool", "organization"})

	metricPoolRunnersDesired = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "pool_runners_desired",
		Namespace: namespace,
		Help:      "Desired number of runners in a pool (replicas)",
	}, []string{"pool", "organization"})

	metricPoolScaleRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "pool_scale_requests_total",
		Namespace: namespace,
		Help:      "Number of scale requests for a pool",
	}, []string{"pool", "organization"})

	metricScaleOperations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "scale_operations_total",
		Namespace: namespace,
		Help:      "Total number of scale operations",
	}, []string{"pool", "organization", "direction", "status"})

	metricScaleDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "scale_duration_seconds",
		Namespace: namespace,
		Help:      "Time taken to complete a scale operation",
		Buckets:   prometheus.DefBuckets,
	}, []string{"pool", "organization", "direction"})

	metricPoolStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "pool_status",
		Namespace: namespace,
		Help:      "Status of a pool. 0 is paused, 1 is active.",
	}, []string{"pool"})
)
