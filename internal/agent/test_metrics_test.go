package agent

import (
	"context"
	"errors"
	"io"
	"math"
	"syscall"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBucketIndex(t *testing.T) {
	cases := []struct {
		ms   float64
		want int
	}{
		{0, 0},
		{5, 0},
		{5.1, 1},
		{10, 1},
		{11, 2},
		{10000, len(latencyBucketsMs) - 1},
		{10001, len(latencyBucketsMs)},
	}

	for _, tc := range cases {
		if got := bucketIndex(tc.ms); got != tc.want {
			t.Errorf("bucketIndex(%v) = %d, want %d", tc.ms, got, tc.want)
		}
	}
}

func TestClassifyHTTPOutcome(t *testing.T) {
	cases := []struct {
		status  int
		outcome string
		code    string
	}{
		{200, outcomeSuccess, ""},
		{204, outcomeSuccess, ""},
		{301, outcomeOther, "301"},
		{404, outcomeClient, "404"},
		{429, outcomeClient, "429"},
		{502, outcomeServer, "502"},
		{504, outcomeServer, "504"},
	}

	for _, tc := range cases {
		outcome, code := classifyHTTPOutcome(tc.status)
		if outcome != tc.outcome || code != tc.code {
			t.Errorf("classifyHTTPOutcome(%d) = (%q, %q), want (%q, %q)",
				tc.status, outcome, code, tc.outcome, tc.code)
		}
	}
}

func TestClassifyRequestError(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		outcome string
	}{
		{"nil", nil, outcomeSuccess},
		{"deadline", context.DeadlineExceeded, outcomeTimeout},
		{"reset", syscall.ECONNRESET, outcomeReset},
		{"broken pipe", syscall.EPIPE, outcomeReset},
		{"refused", syscall.ECONNREFUSED, outcomeConnection},
		{"unreachable", syscall.ENETUNREACH, outcomeConnection},
		{"unexpected eof", io.ErrUnexpectedEOF, outcomeConnection},
		{"canceled", context.Canceled, outcomeOther},
		// Go's HTTP stack frequently returns these as plain strings rather than
		// wrapped sentinels.
		{"dial refused", errors.New("dial tcp 10.0.0.1:80: connect: connection refused"), outcomeConnection},
		{"read reset", errors.New("read tcp: connection reset by peer"), outcomeReset},
		{"goaway", errors.New("http2: server sent GOAWAY and closed the connection"), outcomeConnection},
		{"generic", errors.New("boom"), outcomeOther},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outcome, _ := classifyRequestError(tc.err)
			if outcome != tc.outcome {
				t.Fatalf("outcome = %q, want %q", outcome, tc.outcome)
			}
		})
	}
}

func TestClassifyGRPCError(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		outcome string
	}{
		{"ok", nil, outcomeSuccess},
		{"bad request", status.Error(codes.InvalidArgument, "bad request"), outcomeClient},
		{"deadline", status.Error(codes.DeadlineExceeded, "too slow"), outcomeTimeout},
		{"overloaded", status.Error(codes.Unavailable, "upstream overloaded"), outcomeServer},
		{"refused", status.Error(codes.Unavailable, "connection refused"), outcomeConnection},
		{"closing", status.Error(codes.Unavailable, "transport is closing"), outcomeConnection},
		{"internal", status.Error(codes.Internal, "boom"), outcomeServer},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outcome, _ := classifyGRPCError(tc.err)
			if outcome != tc.outcome {
				t.Fatalf("outcome = %q, want %q", outcome, tc.outcome)
			}
		})
	}
}

func TestTestMetricsRecordsConnectionErrors(t *testing.T) {
	metrics := newTestMetrics(1)

	metrics.RecordRequest(outcomeConnection, "econnrefused", 3, 0, 0)
	metrics.RecordRequest(outcomeConnection, "eof", 3, 0, 0)
	metrics.RecordRequest(outcomeSuccess, "", 3, 0, 0)

	report := metrics.Snapshot()

	if report.ConnectionErrors != 2 {
		t.Errorf("ConnectionErrors = %d, want 2", report.ConnectionErrors)
	}
	if report.SuccessRequests != 1 {
		t.Errorf("SuccessRequests = %d, want 1", report.SuccessRequests)
	}
	if report.OtherErrors != 0 {
		t.Errorf("OtherErrors = %d, want 0", report.OtherErrors)
	}
	if report.ErrorsByCode["econnrefused"] != 1 || report.ErrorsByCode["eof"] != 1 {
		t.Errorf("connection error demux not recorded: %v", report.ErrorsByCode)
	}
}

func TestTestMetricsSnapshotIsCumulative(t *testing.T) {
	metrics := newTestMetrics(3)

	metrics.RecordRequest(outcomeSuccess, "", 12, 100, 10)
	metrics.RecordRequest(outcomeServer, "502", 640, 20, 10)
	metrics.RecordRequest(outcomeTimeout, "timeout", 1200, 0, 5)
	metrics.RecordRequest(outcomeReset, "econnreset", 30, 0, 5)
	metrics.RecordRequest(outcomeClient, "404", 8, 10, 10)

	report := metrics.Snapshot()

	if report.TotalRequests != 5 {
		t.Errorf("TotalRequests = %d, want 5", report.TotalRequests)
	}
	if report.SuccessRequests != 1 {
		t.Errorf("SuccessRequests = %d, want 1", report.SuccessRequests)
	}
	if report.ClientErrors != 1 {
		t.Errorf("ClientErrors = %d, want 1", report.ClientErrors)
	}
	if report.ServerErrors != 1 {
		t.Errorf("ServerErrors = %d, want 1", report.ServerErrors)
	}
	if report.Timeouts != 1 {
		t.Errorf("Timeouts = %d, want 1", report.Timeouts)
	}
	if report.ConnectionResets != 1 {
		t.Errorf("ConnectionResets = %d, want 1", report.ConnectionResets)
	}
	if report.BytesIn != 130 {
		t.Errorf("BytesIn = %d, want 130", report.BytesIn)
	}
	if report.BytesOut != 40 {
		t.Errorf("BytesOut = %d, want 40", report.BytesOut)
	}
	if report.ActiveUsers != 3 {
		t.Errorf("ActiveUsers = %d, want 3", report.ActiveUsers)
	}
	if report.ErrorsByCode["502"] != 1 {
		t.Errorf("ErrorsByCode[502] = %d, want 1", report.ErrorsByCode["502"])
	}

	if report.Latency == nil {
		t.Fatal("Latency histogram is nil")
	}
	if len(report.Latency.Counts) != len(latencyBucketsMs)+1 {
		t.Fatalf("len(Counts) = %d, want %d", len(report.Latency.Counts), len(latencyBucketsMs)+1)
	}
	// Counts are cumulative, so the final bucket must equal the total.
	if last := report.Latency.Counts[len(report.Latency.Counts)-1]; last != 5 {
		t.Errorf("cumulative histogram total = %d, want 5", last)
	}
	if report.Latency.Total != 5 {
		t.Errorf("Latency.Total = %d, want 5", report.Latency.Total)
	}

	// A second snapshot must include the first batch as well.
	metrics.RecordRequest(outcomeSuccess, "", 15, 0, 0)
	second := metrics.Snapshot()

	if second.TotalRequests != 6 {
		t.Errorf("second TotalRequests = %d, want 6", second.TotalRequests)
	}
	if second.Latency.Counts[len(second.Latency.Counts)-1] != 6 {
		t.Errorf("second cumulative histogram total = %d, want 6",
			second.Latency.Counts[len(second.Latency.Counts)-1])
	}
}

func TestProcessSamplerReturnsSaneValues(t *testing.T) {
	sampler := NewProcessSampler()

	cpu, memory := sampler.Sample()
	if cpu != 0 {
		t.Errorf("first CPU sample = %v, want 0", cpu)
	}
	if memory <= 0 {
		t.Errorf("memory = %d, want > 0", memory)
	}

	// Burn a little CPU so the second sample has a measurable window.
	deadline := time.Now().Add(20 * time.Millisecond)
	for time.Now().Before(deadline) {
		_ = math.Sqrt(2)
	}

	cpu, memory = sampler.Sample()
	if cpu < 0 {
		t.Errorf("second CPU sample = %v, want >= 0", cpu)
	}
	if memory <= 0 {
		t.Errorf("second memory = %d, want > 0", memory)
	}
}

func TestCumulativeCPUTimeIsMonotonic(t *testing.T) {
	before := cumulativeCPUTime()

	deadline := time.Now().Add(20 * time.Millisecond)
	for time.Now().Before(deadline) {
		_ = math.Sqrt(3)
	}

	if after := cumulativeCPUTime(); after < before {
		t.Fatalf("cumulative CPU time went backwards: %v -> %v", before, after)
	}
}
