package agent

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Request duration histogram
	AgentRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_agent_request_duration_seconds",
			Help:    "Request duration in seconds from agent perspective",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"target_protocol", "use_case", "method", "status_code", "connection_pooled"},
	)

	// Request counter
	AgentRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_agent_requests_total",
			Help: "Total number of requests made by agent",
		},
		[]string{"target_protocol", "use_case", "method", "status_code", "connection_pooled"},
	)

	// Error counter
	AgentErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_agent_errors_total",
			Help: "Total number of errors encountered by agent",
		},
		[]string{"target_protocol", "use_case", "error_type", "connection_pooled"},
	)

	// Connection time
	AgentConnectionTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_agent_connection_time_seconds",
			Help:    "Connection establishment time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"target_protocol", "use_case", "connection_pooled"},
	)

	// Response time (time to first byte)
	AgentResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_agent_response_time_seconds",
			Help:    "Time to receive first byte of response",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"target_protocol", "use_case", "connection_pooled"},
	)

	// Processing time (total time including reading response body)
	AgentProcessingTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chaos_agent_processing_time_seconds",
			Help:    "Total processing time including response body read",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"target_protocol", "use_case", "connection_pooled"},
	)

	// Success counter
	AgentSuccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chaos_agent_success_total",
			Help: "Total number of successful requests",
		},
		[]string{"target_protocol", "use_case", "connection_pooled"},
	)

	// Active test executions
	AgentActiveTests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "chaos_agent_active_tests",
			Help: "Number of currently active test executions",
		},
	)

	// Active simulated users
	AgentActiveUsers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "chaos_agent_active_users",
			Help: "Number of active simulated users per test",
		},
		[]string{"test_execution_id"},
	)
)
