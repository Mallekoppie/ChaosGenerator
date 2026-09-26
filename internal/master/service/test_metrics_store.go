package service

import (
	"sort"
	"sync"
	"time"

	"mallekoppie/ChaosGenerator/internal/contracts"
)

const (
	// metricsSampleInterval is the minimum spacing between two fleet time-series
	// samples. Agents report roughly once a second and several can report in the
	// same tick, so samples are rate-limited to keep the series readable and the
	// response small.
	metricsSampleInterval = 1 * time.Second
	// metricsMaxSamples bounds the per-execution ring buffer (~5 minutes at one
	// sample per second).
	metricsMaxSamples = 300
	// metricsMaxFinishedRuns is how many completed executions keep their metrics
	// available for the drilldown before the oldest is evicted.
	metricsMaxFinishedRuns = 10
)

// TestMetricsStore aggregates the cumulative metrics reports agents push over
// their stream. It is the master's own telemetry path and does not depend on
// Prometheus or on scraping the agents.
type TestMetricsStore struct {
	mu    sync.RWMutex
	tests map[string]*testMetricsRecord
}

var (
	metricsStore     *TestMetricsStore
	metricsStoreOnce sync.Once
)

// GetTestMetricsStore returns the singleton test metrics store.
func GetTestMetricsStore() *TestMetricsStore {
	metricsStoreOnce.Do(func() {
		metricsStore = &TestMetricsStore{
			tests: make(map[string]*testMetricsRecord),
		}
	})
	return metricsStore
}

// agentMetricsState is the latest cumulative report from one agent plus the
// rates derived from the delta against its previous report.
type agentMetricsState struct {
	agentId    string
	lastUpdate time.Time
	report     *contracts.TestMetricsReport
	previous   *contracts.LatencyHistogram
	egressRps  float64
	errorRate  float64
}

// fleetTotals is the fleet-wide sum of the latest per-agent cumulative counters.
type fleetTotals struct {
	total   int64
	success int64
	client  int64
	server  int64
	timeout int64
	reset   int64
	other   int64
	at      time.Time
}

type testMetricsRecord struct {
	ExecutionId            string
	UseCaseId              string
	UseCaseName            string
	TargetId               string
	TargetName             string
	TargetAddress          string
	TargetProtocol         string
	SimulatedUsersPerAgent int32
	AgentIds               []string
	StartTime              time.Time
	Running                bool
	finishedAt             time.Time

	agents  map[string]*agentMetricsState
	samples []*contracts.TestMetricsSample

	lastSampleAt  time.Time
	lastTotals    fleetTotals
	hasLastTotals bool
}

func newTestMetricsRecord(executionId string) *testMetricsRecord {
	return &testMetricsRecord{
		ExecutionId: executionId,
		Running:     true,
		StartTime:   time.Now(),
		agents:      make(map[string]*agentMetricsState),
		samples:     make([]*contracts.TestMetricsSample, 0, metricsMaxSamples),
	}
}

// EnsureExecution registers (or refreshes) the metadata for a test execution so
// the drilldown can render before the first metrics report arrives.
func (s *TestMetricsStore) EnsureExecution(execution *RunningTest) {
	if execution == nil || execution.TestExecutionId == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record := s.tests[execution.TestExecutionId]
	if record == nil {
		record = newTestMetricsRecord(execution.TestExecutionId)
		s.tests[execution.TestExecutionId] = record
	}

	record.UseCaseId = execution.UseCaseId
	record.UseCaseName = execution.UseCaseName
	record.TargetId = execution.TargetId
	record.TargetName = execution.TargetName
	record.TargetAddress = execution.TargetAddress
	record.TargetProtocol = execution.TargetProtocol
	record.SimulatedUsersPerAgent = execution.SimulatedUsersPerAgent
	record.AgentIds = append([]string(nil), execution.AgentIds...)
	record.StartTime = execution.StartTime
	record.Running = true
	record.finishedAt = time.Time{}
}

// MarkStopped flags an execution as finished so its metrics stay queryable for
// the drilldown, then evicts the oldest finished runs beyond the retention cap.
func (s *TestMetricsStore) MarkStopped(executionId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.tests[executionId]
	if !ok {
		return
	}

	record.Running = false
	record.finishedAt = time.Now()

	s.evictLocked()
}

// evictLocked drops the oldest finished records beyond metricsMaxFinishedRuns.
// Running executions are never evicted.
func (s *TestMetricsStore) evictLocked() {
	finished := make([]*testMetricsRecord, 0, len(s.tests))
	for _, record := range s.tests {
		if !record.Running {
			finished = append(finished, record)
		}
	}

	if len(finished) <= metricsMaxFinishedRuns {
		return
	}

	sort.Slice(finished, func(i, j int) bool {
		return finished[i].finishedAt.Before(finished[j].finishedAt)
	})

	for _, record := range finished[:len(finished)-metricsMaxFinishedRuns] {
		delete(s.tests, record.ExecutionId)
	}
}

// Ingest merges one cumulative report from an agent into the store. Reports for
// an unknown execution (for example after a master restart) create a record so
// telemetry is never silently dropped.
func (s *TestMetricsStore) Ingest(agentId string, report *contracts.TestMetricsReport) {
	if report == nil || report.TestExecutionId == "" {
		return
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	record := s.tests[report.TestExecutionId]
	if record == nil {
		record = newTestMetricsRecord(report.TestExecutionId)
		s.tests[report.TestExecutionId] = record
	}

	record.ingestAgent(agentId, report, now)
	record.appendSample(now)
}

// ingestAgent folds a report into the per-agent state and recomputes the rates
// from the delta against the previous report.
func (r *testMetricsRecord) ingestAgent(agentId string, report *contracts.TestMetricsReport, now time.Time) {
	state := r.agents[agentId]
	if state == nil {
		state = &agentMetricsState{agentId: agentId}
		r.agents[agentId] = state
	}

	if state.report != nil {
		state.egressRps = deltaRate(state.report.TotalRequests, report.TotalRequests, state.lastUpdate, now)
		state.previous = state.report.Latency
	}

	if total := report.TotalRequests; total > 0 {
		state.errorRate = float64(total-report.SuccessRequests) / float64(total)
	}

	state.report = report
	state.lastUpdate = now
}

// deltaRate returns the per-second rate of a monotonically increasing counter
// between two readings.
func deltaRate(previous, current int64, previousAt, now time.Time) float64 {
	elapsed := now.Sub(previousAt).Seconds()
	if elapsed <= 0 {
		return 0
	}

	delta := current - previous
	if delta < 0 {
		delta = 0
	}

	return float64(delta) / elapsed
}

// appendSample records a fleet time-series point, rate-limited to one sample per
// metricsSampleInterval.
func (r *testMetricsRecord) appendSample(now time.Time) {
	if !r.lastSampleAt.IsZero() && now.Sub(r.lastSampleAt) < metricsSampleInterval {
		return
	}

	totals := r.fleetTotalsLocked()

	sample := &contracts.TestMetricsSample{TimestampMs: now.UnixMilli()}

	if r.hasLastTotals {
		sample.EgressRps = deltaRate(r.lastTotals.total, totals.total, r.lastTotals.at, now)

		sample.Ok = max(totals.success-r.lastTotals.success, 0)
		sample.ClientErrors = max(totals.client-r.lastTotals.client, 0)
		sample.ServerErrors = max(totals.server-r.lastTotals.server, 0)
		sample.Timeouts = max(totals.timeout-r.lastTotals.timeout, 0)
		sample.Resets = max(totals.reset-r.lastTotals.reset, 0)

		interval := sample.Ok + sample.ClientErrors + sample.ServerErrors + sample.Timeouts + sample.Resets
		if interval > 0 {
			failures := interval - sample.Ok
			sample.ErrorRate = float64(failures) / float64(interval)
		}

		p50, p90, p95, p99 := quantilesFromHistogram(r.fleetIntervalHistogramLocked())
		sample.P50Ms, sample.P90Ms, sample.P95Ms, sample.P99Ms = p50, p90, p95, p99
	}

	r.samples = append(r.samples, sample)
	if len(r.samples) > metricsMaxSamples {
		r.samples = r.samples[len(r.samples)-metricsMaxSamples:]
	}

	r.lastTotals = totals
	r.hasLastTotals = true
	r.lastSampleAt = now
}

// fleetTotalsLocked sums the latest cumulative counters across every agent that
// has reported.
func (r *testMetricsRecord) fleetTotalsLocked() fleetTotals {
	totals := fleetTotals{at: time.Now()}

	for _, state := range r.agents {
		report := state.report
		if report == nil {
			continue
		}

		totals.total += report.TotalRequests
		totals.success += report.SuccessRequests
		totals.client += report.ClientErrors
		totals.server += report.ServerErrors
		totals.timeout += report.Timeouts
		totals.reset += report.ConnectionResets
		totals.other += report.OtherErrors
	}

	return totals
}

// fleetCumulativeHistogramLocked merges the cumulative latency histograms of
// every agent. Cumulative counts are additive, so summing them yields an exact
// fleet-wide histogram since the start of the test.
func (r *testMetricsRecord) fleetCumulativeHistogramLocked() *contracts.LatencyHistogram {
	merged := &contracts.LatencyHistogram{}

	for _, state := range r.agents {
		report := state.report
		if report == nil || report.Latency == nil {
			continue
		}

		if merged.BoundsMs == nil {
			merged.BoundsMs = report.Latency.BoundsMs
			merged.Counts = make([]int64, len(report.Latency.Counts))
		}

		if len(report.Latency.Counts) != len(merged.Counts) {
			continue
		}

		for i, count := range report.Latency.Counts {
			merged.Counts[i] += count
		}

		merged.Total += report.Latency.Total
		merged.SumMs += report.Latency.SumMs
	}

	if merged.Counts == nil {
		return nil
	}

	return merged
}

// fleetIntervalHistogramLocked builds a histogram covering only the most recent
// report interval per agent, so the time-series curve reflects current traffic
// rather than the whole run. Each agent contributes the delta between its two
// latest cumulative histograms; agents that have reported only once contribute
// their single cumulative histogram.
func (r *testMetricsRecord) fleetIntervalHistogramLocked() *contracts.LatencyHistogram {
	merged := &contracts.LatencyHistogram{}

	for _, state := range r.agents {
		report := state.report
		if report == nil || report.Latency == nil {
			continue
		}

		if merged.BoundsMs == nil {
			merged.BoundsMs = report.Latency.BoundsMs
			merged.Counts = make([]int64, len(report.Latency.Counts))
		}

		if len(report.Latency.Counts) != len(merged.Counts) {
			continue
		}

		previous := state.previous
		previousValid := previous != nil && len(previous.Counts) == len(report.Latency.Counts)

		for i, count := range report.Latency.Counts {
			delta := count
			if previousValid {
				delta -= previous.Counts[i]
			}
			merged.Counts[i] += max(delta, 0)
		}
	}

	if merged.Counts == nil {
		return nil
	}

	// The per-agent deltas are per-bucket counts; quantile() expects cumulative
	// counts, so fold them here.
	var running int64
	for i, count := range merged.Counts {
		running += count
		merged.Counts[i] = running
	}

	return merged
}

// BuildResponse renders the drilldown payload for one execution.
func (s *TestMetricsStore) BuildResponse(executionId string) (*contracts.GetTestMetricsResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.tests[executionId]
	if !ok {
		return nil, false
	}

	connMgr := GetConnectionManager()

	return &contracts.GetTestMetricsResponse{
		Summary:      record.summaryLocked(connMgr),
		Samples:      append([]*contracts.TestMetricsSample(nil), record.samples...),
		Workers:      record.workersLocked(connMgr),
		FleetLatency: record.fleetCumulativeHistogramLocked(),
	}, true
}

// summaryLocked renders the fleet-level KPI envelope for the drilldown.
func (r *testMetricsRecord) summaryLocked(connMgr *ConnectionManager) *contracts.TestMetricsSummary {
	summary := &contracts.TestMetricsSummary{
		TestExecutionId:        r.ExecutionId,
		UseCaseId:              r.UseCaseId,
		UseCaseName:            r.UseCaseName,
		TargetId:               r.TargetId,
		TargetName:             r.TargetName,
		TargetAddress:          r.TargetAddress,
		TargetProtocol:         r.TargetProtocol,
		SimulatedUsersPerAgent: r.SimulatedUsersPerAgent,
		NumberOfAgents:         int32(len(r.AgentIds)),
		StartTime:              r.StartTime.Unix(),
		Running:                r.Running,
	}

	var totalRequests, totalErrors int64
	for _, state := range r.agents {
		if state.report == nil {
			continue
		}

		summary.EgressRps += state.egressRps
		summary.ActiveUsers += int64(state.report.ActiveUsers)
	}

	totals := r.fleetTotalsLocked()
	totalRequests = totals.total
	totalErrors = totals.total - totals.success
	if totalRequests > 0 {
		summary.ErrorRate = float64(totalErrors) / float64(totalRequests)
	}

	summary.TotalRequests = totalRequests
	summary.TotalErrors = totalErrors

	summary.SuccessRequests = totals.success
	summary.ClientErrors = totals.client
	summary.ServerErrors = totals.server
	summary.Timeouts = totals.timeout
	summary.ConnectionResets = totals.reset
	summary.ErrorsByCode = r.mergedErrorsByCodeLocked()

	if summary.ActiveUsers == 0 {
		summary.ActiveUsers = int64(r.SimulatedUsersPerAgent) * int64(len(r.AgentIds))
	}

	summary.P50Ms, summary.P90Ms, summary.P95Ms, summary.P99Ms = quantilesFromHistogram(r.fleetCumulativeHistogramLocked())

	// connectedAgents counts the agents that are both expected for this run and
	// still holding a live stream to the master.
	for _, agentId := range r.AgentIds {
		if connMgr != nil && connMgr.IsAgentConnected(agentId) {
			summary.ConnectedAgents++
		}
	}

	return summary
}

// mergedErrorsByCodeLocked sums the per-agent failure demux counters so the UI
// can break failures down by status code or gRPC code.
func (r *testMetricsRecord) mergedErrorsByCodeLocked() map[string]int64 {
	merged := make(map[string]int64)

	for _, state := range r.agents {
		if state.report == nil {
			continue
		}

		for code, count := range state.report.ErrorsByCode {
			merged[code] += count
		}
	}

	if len(merged) == 0 {
		return nil
	}

	return merged
}

// workersLocked renders one row per agent for the worker fleet table. Agents are
// ordered by their position in the execution so the table is stable across polls.
func (r *testMetricsRecord) workersLocked(connMgr *ConnectionManager) []*contracts.TestMetricsWorker {
	seen := make(map[string]bool, len(r.AgentIds))
	ordered := make([]string, 0, len(r.agents)+len(r.AgentIds))

	for _, agentId := range r.AgentIds {
		if !seen[agentId] {
			seen[agentId] = true
			ordered = append(ordered, agentId)
		}
	}

	// Include agents that reported without being expected (for example an agent
	// that picked up the test after a reconnect).
	extra := make([]string, 0, len(r.agents))
	for agentId := range r.agents {
		if !seen[agentId] {
			extra = append(extra, agentId)
		}
	}
	sort.Strings(extra)
	ordered = append(ordered, extra...)

	workers := make([]*contracts.TestMetricsWorker, 0, len(ordered))
	for _, agentId := range ordered {
		worker := &contracts.TestMetricsWorker{
			AgentId: agentId,
			Online:  connMgr != nil && connMgr.IsAgentConnected(agentId),
		}

		if state, ok := r.agents[agentId]; ok && state.report != nil {
			report := state.report
			worker.ActiveUsers = report.ActiveUsers
			worker.EgressRps = state.egressRps
			worker.ErrorRate = state.errorRate
			worker.CpuPercent = report.CpuPercent
			worker.MemoryBytes = report.MemoryBytes
			worker.MetricsScrapes = report.MetricsScrapes
			worker.LastUpdateMs = state.lastUpdate.UnixMilli()
			worker.ObservedP99Ms = quantile(report.Latency, 0.99)
		}

		workers = append(workers, worker)
	}

	return workers
}

// quantilesFromHistogram returns the p50/p90/p95/p99 latencies in milliseconds
// for a histogram, or zeroes when no observations exist.
func quantilesFromHistogram(histogram *contracts.LatencyHistogram) (p50, p90, p95, p99 float64) {
	if histogram == nil {
		return 0, 0, 0, 0
	}

	return quantile(histogram, 0.50), quantile(histogram, 0.90), quantile(histogram, 0.95), quantile(histogram, 0.99)
}

// quantile interpolates a latency quantile from cumulative histogram buckets.
// Values in the +Inf overflow bucket are capped at the last finite bound.
func quantile(histogram *contracts.LatencyHistogram, q float64) float64 {
	if histogram == nil || len(histogram.Counts) == 0 {
		return 0
	}

	total := histogram.Counts[len(histogram.Counts)-1]
	if total <= 0 {
		return 0
	}

	rank := q * float64(total)

	previousCount := float64(0)
	previousBound := float64(0)

	for i, cumulative := range histogram.Counts {
		if float64(cumulative) < rank {
			previousCount = float64(cumulative)
			if i < len(histogram.BoundsMs) {
				previousBound = histogram.BoundsMs[i]
			}
			continue
		}

		if i >= len(histogram.BoundsMs) {
			return previousBound
		}

		upper := histogram.BoundsMs[i]
		bucketCount := float64(cumulative) - previousCount
		if bucketCount <= 0 {
			return upper
		}

		fraction := (rank - previousCount) / bucketCount
		return previousBound + fraction*(upper-previousBound)
	}

	return previousBound
}
