package metrics

import "github.com/prometheus/client_golang/prometheus"

// ProbesTotal tracks total probe operations partitioned by target, protocol type, and status.
var ProbesTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "watchtower_probes_total",
		Help: "Total number of probe executions by target, type, and outcome status.",
	},
	[]string{"target", "type", "status"},
)

// Legacy alias for backwards compatibility
var ChecksTotal = ProbesTotal

// ProbeDuration tracks probe latency distributions using fine-grained second buckets.
var ProbeDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "watchtower_probe_duration_seconds",
		Help:    "Probe execution latency duration in seconds.",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
	},
	[]string{"target", "type"},
)

// Legacy alias for backwards compatibility
var CheckLatency = ProbeDuration

// TargetUp indicates availability status: 1 = UP, 0 = DOWN / DEGRADED.
var TargetUp = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "watchtower_target_up",
		Help: "Current operational availability status (1 = UP, 0 = DOWN or DEGRADED).",
	},
	[]string{"target", "type"},
)

// SSLCertExpiryDays tracks remaining days before SSL/TLS certificate expiration.
var SSLCertExpiryDays = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "watchtower_ssl_cert_expiry_days",
		Help: "Remaining days before the target SSL/TLS certificate expires.",
	},
	[]string{"target", "type"},
)

// ConsecutiveFailures tracks the current failure streak per target.
var ConsecutiveFailures = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "watchtower_consecutive_failures",
		Help: "Current number of consecutive check failures.",
	},
	[]string{"target", "type"},
)

// ActiveWorkers tracks current in-flight workers.
var ActiveWorkers = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "watchtower_workers_active",
		Help: "Number of worker goroutines currently executing probe checks.",
	},
)

// QueueSize tracks the current buffered depth of the worker job channel.
var QueueSize = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "watchtower_job_queue_size",
		Help: "Current number of queued jobs awaiting worker pickup.",
	},
)

// JobsDroppedTotal records probes dropped due to queue saturation.
var JobsDroppedTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "watchtower_job_dropped_total",
		Help: "Total number of probe jobs dropped due to worker queue saturation.",
	},
	[]string{"target"},
)

// AlertsSentTotal tracks alert dispatch outcomes across notification channels.
var AlertsSentTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "watchtower_alerts_sent_total",
		Help: "Total number of alert notifications attempted by dispatcher and status.",
	},
	[]string{"dispatcher", "status"},
)

// ConfigReloadsTotal tracks configuration hot-reload operations.
var ConfigReloadsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "watchtower_config_reloads_total",
		Help: "Total configuration reload attempts and outcomes.",
	},
	[]string{"status"},
)

func Init() {
	prometheus.MustRegister(
		ProbesTotal,
		ProbeDuration,
		TargetUp,
		SSLCertExpiryDays,
		ConsecutiveFailures,
		ActiveWorkers,
		QueueSize,
		JobsDroppedTotal,
		AlertsSentTotal,
		ConfigReloadsTotal,
	)
}