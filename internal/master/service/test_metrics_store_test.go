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
		ConnectionErrors: 4,
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
	if summary.ConnectionErrors != 4 {
		t.Errorf("ConnectionErrors = %d, want 4", summary.ConnectionErrors)
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

	second := testReport("test-1", 20, 20, &pb.LatencyHistogram{
		BoundsMs: []float64{100, 200, 300},
		Counts:   []int64{10, 10, 20, 20},
		Total:    20,
	})
	second.ConnectionErrors = 3
	store.Ingest("agent-a", second)

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
	if last.ConnectionErrors != 3 {
		t.Errorf("sample ConnectionErrors = %d, want 3", last.ConnectionErrors)
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

func TestStoreStaleAgentRateDecays(t *testing.T) {
	store := newTestStore()

	store.EnsureExecution(&RunningTest{
		TestExecutionId:        "test-1",
		SimulatedUsersPerAgent: 5,
		AgentIds:               []string{"agent-a"},
		StartTime:              time.Now(),
	})

	store.Ingest("agent-a", testReport("test-1", 100, 100, nil))

	// Make the agent look like it was producing 42 req/s but stopped reporting.
	store.mu.Lock()
	state := store.tests["test-1"].agents["agent-a"]
	state.egressRps = 42
	state.lastUpdate = time.Now().Add(-2 * agentMetricsStaleAfter)
	store.mu.Unlock()

	response, found := store.BuildResponse("test-1")
	if !found {
		t.Fatal("BuildResponse did not find the execution")
	}

	// A dead agent's last known rate must not keep looking live.
	if response.Summary.EgressRps != 0 {
		t.Errorf("stale fleet EgressRps = %v, want 0", response.Summary.EgressRps)
	}
	if response.Summary.SilentAgents != 1 {
		t.Errorf("SilentAgents = %d, want 1", response.Summary.SilentAgents)
	}
	if response.Summary.ActiveUsers != 0 {
		t.Errorf("stale ActiveUsers = %d, want 0", response.Summary.ActiveUsers)
	}
	// Cumulative counters must survive: only rates decay.
	if response.Summary.TotalRequests != 100 {
		t.Errorf("TotalRequests = %d, want 100", response.Summary.TotalRequests)
	}

	if len(response.Workers) != 1 {
		t.Fatalf("len(Workers) = %d, want 1", len(response.Workers))
	}
	worker := response.Workers[0]
	if !worker.Stale {
		t.Error("worker should be flagged stale")
	}
	if worker.EgressRps != 0 {
		t.Errorf("stale worker EgressRps = %v, want 0", worker.EgressRps)
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

func TestStoreWorkerReportsCumulativeMetrics(t *testing.T) {
	store := newTestStore()

	store.EnsureExecution(&RunningTest{
		TestExecutionId: "test-1",
		AgentIds:        []string{"agent-a"},
		StartTime:       time.Now(),
	})

	store.Ingest("agent-a", &pb.TestMetricsReport{
		TestExecutionId:  "test-1",
		TotalRequests:    100,
		SuccessRequests:  90,
		ClientErrors:     4,
		ServerErrors:     3,
		Timeouts:         1,
		ConnectionResets: 1,
		OtherErrors:      1,
		ConnectionErrors: 2,
		BytesIn:          1000,
		BytesOut:         2000,
		Latency: &pb.LatencyHistogram{
			BoundsMs: []float64{100, 200},
			Counts:   []int64{0, 10, 100},
			Total:    100,
			SumMs:    5000,
		},
	})

	response, found := store.BuildResponse("test-1")
	if !found {
		t.Fatal("BuildResponse did not find the execution")
	}
	if len(response.Workers) != 1 {
		t.Fatalf("len(Workers) = %d, want 1", len(response.Workers))
	}

	worker := response.Workers[0]
	if worker.TotalRequests != 100 || worker.SuccessRequests != 90 {
		t.Errorf("worker requests = %d/%d, want 100/90", worker.TotalRequests, worker.SuccessRequests)
	}
	if worker.ClientErrors != 4 || worker.ServerErrors != 3 {
		t.Errorf("worker client/server errors = %d/%d, want 4/3", worker.ClientErrors, worker.ServerErrors)
	}
	if worker.Timeouts != 1 || worker.ConnectionResets != 1 || worker.OtherErrors != 1 {
		t.Errorf("worker timeouts/resets/other = %d/%d/%d, want 1/1/1",
			worker.Timeouts, worker.ConnectionResets, worker.OtherErrors)
	}
	if worker.ConnectionErrors != 2 {
		t.Errorf("worker connection errors = %d, want 2", worker.ConnectionErrors)
	}
	if worker.BytesIn != 1000 || worker.BytesOut != 2000 {
		t.Errorf("worker bytes = %d/%d, want 1000/2000", worker.BytesIn, worker.BytesOut)
	}
	// 5000ms of latency across 100 observations averages 50ms.
	if math.Abs(worker.AvgLatencyMs-50) > 0.001 {
		t.Errorf("worker AvgLatencyMs = %v, want 50", worker.AvgLatencyMs)
	}
}

func TestStoreSummaryEndTimeAndDuration(t *testing.T) {
	store := newTestStore()

	store.EnsureExecution(&RunningTest{
		TestExecutionId: "test-1",
		StartTime:       time.Now().Add(-2 * time.Second),
	})
	store.Ingest("agent-a", testReport("test-1", 10, 10, nil))

	running, _ := store.BuildResponse("test-1")
	if running.Summary.EndTime != 0 {
		t.Errorf("running EndTime = %d, want 0", running.Summary.EndTime)
	}
	if running.Summary.DurationMs <= 0 {
		t.Errorf("running DurationMs = %d, want > 0", running.Summary.DurationMs)
	}

	store.MarkStopped("test-1")

	stopped, _ := store.BuildResponse("test-1")
	if stopped.Summary.EndTime == 0 {
		t.Error("stopped EndTime = 0, want the finish timestamp")
	}
	if stopped.Summary.DurationMs < 1900 {
		t.Errorf("stopped DurationMs = %d, want >= 1900", stopped.Summary.DurationMs)
	}
}

func TestStoreHistoryNewestFirstAndCapped(t *testing.T) {
	store := newTestStore()

	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("test-%d", i)
		store.EnsureExecution(&RunningTest{TestExecutionId: id, StartTime: time.Now()})
		store.Ingest("agent-a", testReport(id, int64(10*(i+1)), int64(10*(i+1)), nil))
		store.MarkStopped(id)

		// Space out the finish times so ordering is deterministic.
		time.Sleep(time.Millisecond)
	}

	runs := store.History(3)
	if len(runs) != 3 {
		t.Fatalf("len(History(3)) = %d, want 3", len(runs))
	}
	if runs[0].TestExecutionId != "test-4" {
		t.Errorf("newest run = %q, want test-4", runs[0].TestExecutionId)
	}
	if runs[2].TotalRequests != 30 {
		t.Errorf("oldest returned run requests = %d, want 30", runs[2].TotalRequests)
	}

	// A zero limit falls back to the configured retention.
	if got := len(store.History(0)); got != 5 {
		t.Errorf("len(History(0)) = %d, want 5", got)
	}
}

func TestRecordRoundTripsThroughHistoryModel(t *testing.T) {
	store := newTestStore()

	store.EnsureExecution(&RunningTest{
		TestExecutionId:        "test-1",
		UseCaseId:              "uc7",
		UseCaseName:            "Report Generation",
		TargetId:               "target-1",
		TargetName:             "Payments",
		TargetAddress:          "https://payments.internal",
		TargetProtocol:         "https",
		SimulatedUsersPerAgent: 5,
		AgentIds:               []string{"agent-a"},
		StartTime:              time.Now().Add(-time.Second),
	})
	store.Ingest("agent-a", &pb.TestMetricsReport{
		TestExecutionId: "test-1",
		TotalRequests:   100,
		SuccessRequests: 95,
		ServerErrors:    5,
		BytesIn:         111,
		BytesOut:        222,
		Latency:         &pb.LatencyHistogram{BoundsMs: []float64{100}, Counts: []int64{10, 90}, Total: 100, SumMs: 2000},
	})
	store.MarkStopped("test-1")

	store.mu.RLock()
	original := store.tests["test-1"]
	model := original.toHistoryModel()
	store.mu.RUnlock()

	restored := recordFromHistoryModel(model)
	if restored.ExecutionId != "test-1" {
		t.Errorf("restored id = %q, want test-1", restored.ExecutionId)
	}
	if restored.Running {
		t.Error("restored run should be finished")
	}
	if restored.finishedAt.IsZero() {
		t.Error("restored run has no end time")
	}
	if len(restored.samples) != len(original.samples) {
		t.Errorf("restored samples = %d, want %d", len(restored.samples), len(original.samples))
	}

	before := original.summaryLocked(nil)
	after := restored.summaryLocked(nil)

	if before.TotalRequests != after.TotalRequests || before.TotalErrors != after.TotalErrors {
		t.Errorf("totals changed across round-trip: %d/%d -> %d/%d",
			before.TotalRequests, before.TotalErrors, after.TotalRequests, after.TotalErrors)
	}
	if before.BytesIn != after.BytesIn || before.BytesOut != after.BytesOut {
		t.Errorf("bytes changed across round-trip: %d/%d -> %d/%d",
			before.BytesIn, before.BytesOut, after.BytesIn, after.BytesOut)
	}
	if before.EndTime != after.EndTime || before.DurationMs != after.DurationMs {
		t.Errorf("end/duration changed across round-trip: %d/%d -> %d/%d",
			before.EndTime, before.DurationMs, after.EndTime, after.DurationMs)
	}
	if before.NumberOfAgents != after.NumberOfAgents {
		t.Errorf("agent count changed across round-trip: %d -> %d",
			before.NumberOfAgents, after.NumberOfAgents)
	}
}
