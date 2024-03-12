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
		Name:      "up",
		Namespace: namespace,
		Subsystem: "server",
		Help:      "Is the server up",
	})

	metricPoolMaxRunnersCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "max_runners_count",
		Namespace: namespace,
		Subsystem: "pool",
		Help:      "Maximum number of runners in a pool",
	}, []string{"pool"})

	metricPoolMinRunnersCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "min_runners_count",
		Namespace: namespace,
		Subsystem: "pool",
		Help:      "Minimum number of runners in a pool",
	}, []string{"pool"})

	metricPoolCurrentRunnersCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "current_runners_count",
		Namespace: namespace,
		Subsystem: "pool",
		Help:      "Current number of runners in a pool",
	}, []string{"pool"})

	metricPoolScaleRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "scale_requests",
		Namespace: namespace,
		Subsystem: "pool",
		Help:      "Number of scale requests for a pool",
	}, []string{"pool"})

	metricPoolScaleFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "scale_failures",
		Namespace: namespace,
		Subsystem: "pool",
		Help:      "Number of scale failures for a pool",
	}, []string{"pool"})

	metricPoolScaleSuccesses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "scale_successes",
		Namespace: namespace,
		Subsystem: "pool",
		Help:      "Number of scale successes for a pool",
	}, []string{"pool"})

	metricPoolTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name:      "total",
		Namespace: namespace,
		Subsystem: "pool",
		Help:      "Total number of pools",
	})

	metricPoolStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "status",
		Namespace: namespace,
		Subsystem: "pool",
		Help:      "Status of a pool. 0 is paused, 1 is active.",
	}, []string{"pool"})
)
