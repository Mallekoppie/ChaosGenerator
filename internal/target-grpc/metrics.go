package targetgrpc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Request duration histogram
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_target_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"use_case", "method", "status_code"},
	)

	// Request counter
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_target_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"use_case", "method", "status_code"},
	)

	// Processing time
	ProcessingTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_target_grpc_processing_time_seconds",
			Help:    "Request processing time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"use_case"},
	)

	// Error counter
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_target_grpc_errors_total",
			Help: "Total number of errors",
		},
		[]string{"use_case", "error_type"},
	)

	// Success counter
	SuccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_target_grpc_success_total",
			Help: "Total number of successful requests",
		},
		[]string{"use_case"},
	)

	// Response size histogram
	ResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_target_grpc_response_size_bytes",
			Help:    "gRPC response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000, 100000000},
		},
		[]string{"use_case"},
	)

	// Request size histogram
	RequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_target_grpc_request_size_bytes",
			Help:    "gRPC request size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000, 100000000},
		},
		[]string{"use_case"},
	)
)
