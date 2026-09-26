package agent

import (
	"context"
	"errors"
	"io"
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
	outcomeSuccess    = "success"
	outcomeClient     = "client_error"
	outcomeServer     = "server_error"
	outcomeTimeout    = "timeout"
	outcomeReset      = "connection_reset"
	outcomeConnection = "connection_error"
	outcomeOther      = "other_error"
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
	connErrors   int64
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
	case outcomeConnection:
		m.connErrors++
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
		ConnectionErrors: m.connErrors,
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

	// Timeouts are their own class: the connection may be healthy but too slow.
	if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		return outcomeTimeout, "timeout"
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return outcomeTimeout, "timeout"
	}

	if errors.Is(err, context.Canceled) {
		return outcomeOther, "canceled"
	}

	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) {
		return outcomeReset, "econnreset"
	}

	switch {
	case errors.Is(err, syscall.ECONNREFUSED):
		return outcomeConnection, "econnrefused"
	case errors.Is(err, syscall.EHOSTUNREACH):
		return outcomeConnection, "ehostunreach"
	case errors.Is(err, syscall.ENETUNREACH):
		return outcomeConnection, "enetunreach"
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return outcomeConnection, "eof"
	case errors.Is(err, net.ErrClosed):
		return outcomeConnection, "closed"
	}

	if outcome, code := classifyConnectionMessage(err.Error()); outcome != "" {
		return outcome, code
	}

	return outcomeOther, "transport"
}

// classifyConnectionMessage inspects a transport error message for a
// connection-level failure. Go's HTTP and gRPC stacks wrap lower-level errors in
// plain strings on several code paths, so matching the message catches failures
// the sentinel checks miss. It returns empty strings when the message does not
// describe a connection-level failure.
func classifyConnectionMessage(message string) (outcome, code string) {
	lower := strings.ToLower(message)

	switch {
	case strings.Contains(lower, "connection reset"),
		strings.Contains(lower, "broken pipe"),
		strings.Contains(lower, "reset by peer"):
		return outcomeReset, "econnreset"
	case strings.Contains(lower, "connection refused"):
		return outcomeConnection, "econnrefused"
	case strings.Contains(lower, "no route to host"):
		return outcomeConnection, "ehostunreach"
	case strings.Contains(lower, "network is unreachable"):
		return outcomeConnection, "enetunreach"
	case strings.Contains(lower, "server closed idle connection"),
		strings.Contains(lower, "goaway"),
		strings.Contains(lower, "transport is closing"),
		strings.Contains(lower, "error while dialing"),
		strings.Contains(lower, "connection closed"):
		return outcomeConnection, "closed"
	case strings.Contains(lower, "eof"):
		return outcomeConnection, "eof"
	}

	return "", ""
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
	case codes.Unavailable:
		// A failure to reach the target (refused, unreachable, reset, closing) is
		// a transport problem rather than an error response, so keep it out of
		// the server-error bucket.
		if outcome, code := classifyConnectionMessage(st.Message()); outcome != "" {
			return outcome, code
		}
		return outcomeServer, st.Code().String()
	case codes.Internal, codes.Unknown, codes.DataLoss, codes.ResourceExhausted, codes.Aborted:
		return outcomeServer, st.Code().String()
	default:
		return outcomeOther, st.Code().String()
	}
}
