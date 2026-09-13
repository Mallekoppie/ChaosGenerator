package targetgrpc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Request duration histogram
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_target_request_duration_seconds",
			Help:    "Target service request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"target_type", "use_case", "method", "status_code"},
	)

	// Request counter
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_target_requests_total",
			Help: "Total number of target service requests",
		},
		[]string{"target_type", "use_case", "method", "status_code"},
	)

	// Processing time
	ProcessingTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_target_processing_time_seconds",
			Help:    "Request processing time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"target_type", "use_case"},
	)

	// Error counter
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_target_errors_total",
			Help: "Total number of errors",
		},
		[]string{"target_type", "use_case", "error_type"},
	)

	// Success counter
	SuccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_target_success_total",
			Help: "Total number of successful requests",
		},
		[]string{"target_type", "use_case"},
	)

	// Response size histogram
	ResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_target_response_size_bytes",
			Help:    "Target service response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000, 100000000},
		},
		[]string{"target_type", "use_case"},
	)

	// Request size histogram
	RequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_target_request_size_bytes",
			Help:    "Target service request size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000, 100000000},
		},
		[]string{"target_type", "use_case"},
	)
)
