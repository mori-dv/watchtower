package metrics

import "github.com/prometheus/client_golang/prometheus"

var ChecksTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "watchtower_checks_total",
		Help: "Total number of checks",
	},
	[]string{
		"target",
		"type",
		"status",
	},
)

var CheckLatency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "watchtower_check_duration_seconds",
		Help: "Check latency",
		Buckets: prometheus.DefBuckets,
	},
	[]string{
		"target",
		"type",
	},
)

var TargetUp = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "watchtower_target_up",
		Help: "Target status",
	},
	[]string{
		"target",
		"type",
	},
)

var ActiveWorkers = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "watchtower_workers_active",
		Help: "Currently active workers",
	},
)

var QueueSize = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "watchtower_job_queue_size",
		Help: "Current queue size",
	},
)

func Init() {
	prometheus.MustRegister(
		ChecksTotal,
		CheckLatency,
		TargetUp,
		ActiveWorkers,
		QueueSize,
	)
}