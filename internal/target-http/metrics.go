package targethttp

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Request duration histogram
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_target_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"use_case", "method", "status_code"},
	)

	// Request counter
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_target_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"use_case", "method", "status_code"},
	)

	// Connection time
	ConnectionTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_target_http_connection_time_seconds",
			Help:    "Connection establishment time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"use_case"},
	)

	// Processing time
	ProcessingTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_target_http_processing_time_seconds",
			Help:    "Request processing time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"use_case"},
	)

	// Error counter
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_target_http_errors_total",
			Help: "Total number of errors",
		},
		[]string{"use_case", "error_type"},
	)

	// Success counter
	SuccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_target_http_success_total",
			Help: "Total number of successful requests",
		},
		[]string{"use_case"},
	)
)
