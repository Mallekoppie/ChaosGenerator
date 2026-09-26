package service

import (
	"fmt"
	"math"
	"testing"
	"time"

	pb "mallekoppie/ChaosGenerator/internal/contracts"
)

func newTestStore() *TestMetricsStore {
	return &TestMetricsStore{tests: make(map[string]*testMetricsRecord)}
}

func testReport(executionId string, total, success int64, latency *pb.LatencyHistogram) *pb.TestMetricsReport {
	return &pb.TestMetricsReport{
		TestExecutionId: executionId,
		TimestampMs:     time.Now().UnixMilli(),
		TotalRequests:   total,
		SuccessRequests: success,
		ActiveUsers:     4,
		Latency:         latency,
	}
}

func TestDeltaRate(t *testing.T) {
	base := time.Now()

	if got := deltaRate(100, 120, base, base.Add(2*time.Second)); math.Abs(got-10) > 0.001 {
		t.Errorf("deltaRate = %v, want 10", got)
	}
	// A counter that appears to go backwards must not produce a negative rate.
	if got := deltaRate(120, 100, base, base.Add(time.Second)); got != 0 {
		t.Errorf("deltaRate for a regressed counter = %v, want 0", got)
	}
	// A zero-width window must not divide by zero.
	if got := deltaRate(0, 10, base, base); got != 0 {
		t.Errorf("deltaRate over zero elapsed = %v, want 0", got)
	}
}

func TestQuantileFromHistogram(t *testing.T) {
	histogram := &pb.LatencyHistogram{
		BoundsMs: []float64{100, 200},
		Counts:   []int64{10, 10, 10},
		Total:    10,
	}

	cases := []struct {
		q    float64
		want float64
	}{
		{0.50, 50},
		{0.90, 90},
		// All observations sit in the (0, 100] bucket, so the assumption of a
		// uniform spread inside the bucket yields 99 for p99.
		{0.99, 99},
	}

	for _, tc := range cases {
		if got := quantile(histogram, tc.q); math.Abs(got-tc.want) > 0.001 {
			t.Errorf("quantile(%v) = %v, want %v", tc.q, got, tc.want)
		}
	}

	if got := quantile(nil, 0.5); got != 0 {
		t.Errorf("quantile(nil) = %v, want 0", got)
	}
}

func TestStoreSummaryAggregatesAgents(t *testing.T) {
	store := newTestStore()

	store.EnsureExecution(&RunningTest{
		TestExecutionId:        "test-1",
		UseCaseId:              "uc7",
		UseCaseName:            "Report Generation",
		TargetId:               "target-1",
		TargetName:             "Payments",
		TargetAddress:          "https://payments.internal:8443",
		TargetProtocol:         "https",
		SimulatedUsersPerAgent: 10,
		AgentIds:               []string{"agent-a", "agent-b"},
		StartTime:              time.Now(),
	})

	store.Ingest("agent-a", &pb.TestMetricsReport{
		TestExecutionId:  "test-1",
		TotalRequests:    100,
		SuccessRequests:  95,
		ServerErrors:     3,
		ClientErrors:     1,
		Timeouts:         1,
		ConnectionResets: 2,
		ActiveUsers:      10,
		ErrorsByCode:     map[string]int64{"502": 2, "504": 1},
		Latency:          &pb.LatencyHistogram{BoundsMs: []float64{100, 200}, Counts: []int64{10, 10, 100}, Total: 100},
	})

	store.Ingest("agent-b", &pb.TestMetricsReport{
		TestExecutionId: "test-1",
		TotalRequests:   50,
		SuccessRequests: 50,
		ActiveUsers:     10,
		ErrorsByCode:    map[string]int64{"502": 1},
		Latency:         &pb.LatencyHistogram{BoundsMs: []float64{100, 200}, Counts: []int64{5, 5, 50}, Total: 50},
	})

	response, found := store.BuildResponse("test-1")
	if !found {
		t.Fatal("BuildResponse did not find the execution")
	}

	summary := response.Summary
	if summary.TotalRequests != 150 {
		t.Errorf("TotalRequests = %d, want 150", summary.TotalRequests)
	}
	if summary.SuccessRequests != 145 {
		t.Errorf("SuccessRequests = %d, want 145", summary.SuccessRequests)
	}
	if summary.ClientErrors != 1 {
		t.Errorf("ClientErrors = %d, want 1", summary.ClientErrors)
	}
	if summary.ServerErrors != 3 {
		t.Errorf("ServerErrors = %d, want 3", summary.ServerErrors)
	}
	if summary.Timeouts != 1 || summary.ConnectionResets != 2 {
		t.Errorf("timeouts/resets = %d/%d, want 1/2", summary.Timeouts, summary.ConnectionResets)
	}
	if summary.TotalErrors != 5 {
		t.Errorf("TotalErrors = %d, want 5", summary.TotalErrors)
	}
	if math.Abs(summary.ErrorRate-(5.0/150.0)) > 0.0001 {
		t.Errorf("ErrorRate = %v, want %v", summary.ErrorRate, 5.0/150.0)
	}
	if summary.ActiveUsers != 20 {
		t.Errorf("ActiveUsers = %d, want 20", summary.ActiveUsers)
	}
	if summary.TargetAddress != "https://payments.internal:8443" {
		t.Errorf("TargetAddress = %q", summary.TargetAddress)
	}
	if summary.NumberOfAgents != 2 {
		t.Errorf("NumberOfAgents = %d, want 2", summary.NumberOfAgents)
	}
	if summary.ErrorsByCode["502"] != 3 {
		t.Errorf("ErrorsByCode[502] = %d, want 3", summary.ErrorsByCode["502"])
	}
	if len(response.Workers) != 2 {
		t.Fatalf("len(Workers) = %d, want 2", len(response.Workers))
	}
	if response.FleetLatency == nil || response.FleetLatency.Counts[len(response.FleetLatency.Counts)-1] != 150 {
		t.Errorf("fleet cumulative histogram does not total 150")
	}
}

func TestStoreIntervalSampleUsesDeltas(t *testing.T) {
	store := newTestStore()

	record := newTestMetricsRecord("test-1")
	record.StartTime = time.Now().Add(-10 * time.Second)
	store.tests["test-1"] = record

	store.Ingest("agent-a", testReport("test-1", 10, 10, &pb.LatencyHistogram{
		BoundsMs: []float64{100, 200, 300},
		Counts:   []int64{10, 10, 10, 10},
		Total:    10,
	}))

	// Force the next ingest to fall outside the sampling interval and give the
	// rate calculation a deterministic two-second window.
	store.mu.Lock()
	record.lastSampleAt = time.Now().Add(-2 * time.Second)
	record.lastTotals.at = time.Now().Add(-2 * time.Second)
	store.mu.Unlock()

	store.Ingest("agent-a", testReport("test-1", 20, 20, &pb.LatencyHistogram{
		BoundsMs: []float64{100, 200, 300},
		Counts:   []int64{10, 10, 20, 20},
		Total:    20,
	}))

	response, found := store.BuildResponse("test-1")
	if !found {
		t.Fatal("BuildResponse did not find the execution")
	}

	if len(response.Samples) != 2 {
		t.Fatalf("len(Samples) = %d, want 2", len(response.Samples))
	}

	last := response.Samples[1]
	if last.Ok != 10 {
		t.Errorf("sample Ok delta = %d, want 10", last.Ok)
	}
	if math.Abs(last.EgressRps-5) > 0.001 {
		t.Errorf("sample EgressRps = %v, want 5", last.EgressRps)
	}
	// The interval histogram delta is {0, 0, 10, 10}, so latency has shifted
	// into the (200, 300] bucket and p50/p99 are interpolated inside it.
	if last.P50Ms <= 200 || last.P50Ms > 300 {
		t.Errorf("sample P50Ms = %v, want in (200, 300]", last.P50Ms)
	}
	if last.P99Ms < last.P50Ms || last.P99Ms > 300 {
		t.Errorf("sample P99Ms = %v, want in [P50, 300]", last.P99Ms)
	}
}

func TestStoreCreatesRecordForUnknownExecution(t *testing.T) {
	store := newTestStore()

	store.Ingest("agent-a", testReport("test-unknown", 5, 5, nil))

	response, found := store.BuildResponse("test-unknown")
	if !found {
		t.Fatal("Ingest did not create a record for an unknown execution")
	}
	if response.Summary.TotalRequests != 5 {
		t.Errorf("TotalRequests = %d, want 5", response.Summary.TotalRequests)
	}
}

func TestStoreRetentionKeepsRunningAndNewestFinished(t *testing.T) {
	store := newTestStore()

	// A running execution must never be evicted, even past the retention cap.
	store.EnsureExecution(&RunningTest{TestExecutionId: "running", StartTime: time.Now()})

	ids := make([]string, 0, metricsMaxFinishedRuns+3)
	for i := 0; i < metricsMaxFinishedRuns+3; i++ {
		id := fmt.Sprintf("test-%02d", i)
		ids = append(ids, id)

		store.EnsureExecution(&RunningTest{TestExecutionId: id, StartTime: time.Now()})
		store.MarkStopped(id)

		// Space out the finish times so eviction order is deterministic.
		time.Sleep(time.Millisecond)
	}

	store.mu.RLock()
	total := len(store.tests)
	store.mu.RUnlock()

	if total != metricsMaxFinishedRuns+1 {
		t.Fatalf("retained records = %d, want %d", total, metricsMaxFinishedRuns+1)
	}

	if _, found := store.BuildResponse("running"); !found {
		t.Error("running execution was evicted")
	}
	if _, found := store.BuildResponse(ids[len(ids)-1]); !found {
		t.Error("newest finished execution was evicted")
	}
	if _, found := store.BuildResponse(ids[0]); found {
		t.Error("oldest finished execution was not evicted")
	}
}
