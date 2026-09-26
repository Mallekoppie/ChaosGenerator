package agent

import (
	"context"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	pb "mallekoppie/ChaosGenerator/internal/contracts"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// latencyBucketsMs are the inclusive upper bounds of the latency histogram
// buckets in milliseconds. They mirror the Prometheus default buckets scaled to
// milliseconds so the gRPC telemetry the agent reports to the master lines up
// with what it exports on /metrics.
var latencyBucketsMs = []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

// Outcome classes for a completed request.
const (
	outcomeSuccess = "success"
	outcomeClient  = "client_error"
	outcomeServer  = "server_error"
	outcomeTimeout = "timeout"
	outcomeReset   = "connection_reset"
	outcomeOther   = "other_error"
)

// TestMetrics accumulates client-side telemetry for a single test execution.
//
// Every counter is cumulative for the lifetime of the execution, so a report
// lost on the wire self-heals: the master derives rates from the delta between
// consecutive snapshots rather than trusting any single report.
type TestMetrics struct {
	mu sync.Mutex

	total        int64
	success      int64
	clientErrors int64
	serverErrors int64
	timeouts     int64
	resets       int64
	otherErrors  int64
	errorsByCode map[string]int64

	bytesIn  int64
	bytesOut int64

	// buckets holds cumulative histogram counts and always has
	// len(latencyBucketsMs)+1 entries; the last entry is the overflow bucket.
	buckets []int64
	sumMs   float64

	activeUsers int
}

// newTestMetrics creates a collector for a test execution with the given number
// of simulated users.
func newTestMetrics(activeUsers int) *TestMetrics {
	return &TestMetrics{
		buckets:      make([]int64, len(latencyBucketsMs)+1),
		errorsByCode: make(map[string]int64),
		activeUsers:  activeUsers,
	}
}

// RecordRequest records one completed request attempt.
//
// outcome is one of the outcome* constants, code is a stable identifier for the
// failure demux (HTTP status, gRPC code or a synthetic transport label) and is
// ignored for successful requests. durationMs is the end-to-end client-observed
// latency, inBytes/outBytes the payload volume of the exchange.
func (m *TestMetrics) RecordRequest(outcome, code string, durationMs float64, inBytes, outBytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.total++

	switch outcome {
	case outcomeSuccess:
		m.success++
	case outcomeClient:
		m.clientErrors++
	case outcomeServer:
		m.serverErrors++
	case outcomeTimeout:
		m.timeouts++
	case outcomeReset:
		m.resets++
	default:
		m.otherErrors++
	}

	if outcome != outcomeSuccess && code != "" {
		m.errorsByCode[code]++
	}

	if durationMs < 0 {
		durationMs = 0
	}
	m.buckets[bucketIndex(durationMs)]++
	m.sumMs += durationMs
	m.bytesIn += inBytes
	m.bytesOut += outBytes
}

// Snapshot returns a cumulative report for this execution. The caller fills in
// the execution id and process-wide telemetry (CPU, memory, scrape count).
func (m *TestMetrics) Snapshot() *pb.TestMetricsReport {
	m.mu.Lock()
	defer m.mu.Unlock()

	errorsByCode := make(map[string]int64, len(m.errorsByCode))
	for code, count := range m.errorsByCode {
		errorsByCode[code] = count
	}

	// The collector stores one count per bucket; the wire format is cumulative
	// (Prometheus convention) so the master can sum histograms across agents.
	counts := make([]int64, len(m.buckets))
	var cumulative int64
	for i, count := range m.buckets {
		cumulative += count
		counts[i] = cumulative
	}

	return &pb.TestMetricsReport{
		TimestampMs:      time.Now().UnixMilli(),
		TotalRequests:    m.total,
		SuccessRequests:  m.success,
		ClientErrors:     m.clientErrors,
		ServerErrors:     m.serverErrors,
		Timeouts:         m.timeouts,
		ConnectionResets: m.resets,
		OtherErrors:      m.otherErrors,
		ErrorsByCode:     errorsByCode,
		BytesIn:          m.bytesIn,
		BytesOut:         m.bytesOut,
		Latency: &pb.LatencyHistogram{
			BoundsMs: latencyBucketsMs,
			Counts:   counts,
			Total:    m.total,
			SumMs:    m.sumMs,
		},
		ActiveUsers: int32(m.activeUsers),
	}
}

// bucketIndex returns the index of the first bucket whose bound is >= ms, or the
// overflow bucket when ms exceeds every bound.
func bucketIndex(ms float64) int {
	for i, bound := range latencyBucketsMs {
		if ms <= bound {
			return i
		}
	}
	return len(latencyBucketsMs)
}

// classifyHTTPOutcome maps an HTTP status code onto an outcome class and a
// stable demux code.
func classifyHTTPOutcome(statusCode int) (outcome, code string) {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return outcomeSuccess, ""
	case statusCode >= 400 && statusCode < 500:
		return outcomeClient, strconv.Itoa(statusCode)
	case statusCode >= 500:
		return outcomeServer, strconv.Itoa(statusCode)
	default:
		return outcomeOther, strconv.Itoa(statusCode)
	}
}

// classifyRequestError maps a transport-level error onto an outcome class and a
// stable demux code.
func classifyRequestError(err error) (outcome, code string) {
	if err == nil {
		return outcomeSuccess, ""
	}

	if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		return outcomeTimeout, "timeout"
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return outcomeTimeout, "timeout"
	}

	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) {
		return outcomeReset, "econnreset"
	}

	if errors.Is(err, context.Canceled) {
		return outcomeOther, "canceled"
	}

	if strings.Contains(strings.ToLower(err.Error()), "connection reset") {
		return outcomeReset, "econnreset"
	}

	return outcomeOther, "transport"
}

// classifyGRPCError maps a gRPC status onto an outcome class and a stable demux
// code.
func classifyGRPCError(err error) (outcome, code string) {
	if err == nil {
		return outcomeSuccess, ""
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return outcomeTimeout, "DeadlineExceeded"
	}

	st, ok := status.FromError(err)
	if !ok {
		return classifyRequestError(err)
	}

	switch st.Code() {
	case codes.OK:
		return outcomeSuccess, ""
	case codes.Canceled,
		codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.FailedPrecondition,
		codes.OutOfRange,
		codes.Unauthenticated:
		return outcomeClient, st.Code().String()
	case codes.DeadlineExceeded:
		return outcomeTimeout, st.Code().String()
	case codes.Unavailable, codes.Internal, codes.Unknown, codes.DataLoss, codes.ResourceExhausted, codes.Aborted:
		if strings.Contains(strings.ToLower(st.Message()), "connection reset") {
			return outcomeReset, "econnreset"
		}
		return outcomeServer, st.Code().String()
	default:
		return outcomeOther, st.Code().String()
	}
}
