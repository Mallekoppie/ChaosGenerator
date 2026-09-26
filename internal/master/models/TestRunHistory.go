package models

import "time"

// TestRunHistory is the persisted snapshot of one finished test execution.
//
// The live metrics store is in-memory only, so a master restart would otherwise
// lose every run that was not exported. Persisting the fleet-level metadata, the
// latest cumulative report from each agent and the fleet time-series lets the
// history view and the documentation report survive a restart.
type TestRunHistory struct {
	ExecutionId            string              `json:"executionId"`
	UseCaseId              string              `json:"useCaseId"`
	UseCaseName            string              `json:"useCaseName"`
	TargetId               string              `json:"targetId"`
	TargetName             string              `json:"targetName"`
	TargetAddress          string              `json:"targetAddress"`
	TargetProtocol         string              `json:"targetProtocol"`
	SimulatedUsersPerAgent int32               `json:"simulatedUsersPerAgent"`
	AgentIds               []string            `json:"agentIds"`
	StartTime              time.Time           `json:"startTime"`
	EndTime                time.Time           `json:"endTime"`
	Running                bool                `json:"running"`
	Agents                 []TestRunAgentSnapshot `json:"agents"`
	Samples                []TestRunSample     `json:"samples,omitempty"`
}

// TestRunAgentSnapshot is the latest cumulative report one agent produced for a
// run, flattened so the history does not depend on the protobuf package.
type TestRunAgentSnapshot struct {
	AgentId          string           `json:"agentId"`
	ActiveUsers      int32            `json:"activeUsers"`
	TotalRequests    int64            `json:"totalRequests"`
	SuccessRequests  int64            `json:"successRequests"`
	ClientErrors     int64            `json:"clientErrors"`
	ServerErrors     int64            `json:"serverErrors"`
	Timeouts         int64            `json:"timeouts"`
	ConnectionResets int64            `json:"connectionResets"`
	OtherErrors      int64            `json:"otherErrors"`
	ConnectionErrors int64            `json:"connectionErrors"`
	ErrorsByCode     map[string]int64 `json:"errorsByCode,omitempty"`
	BytesIn          int64            `json:"bytesIn"`
	BytesOut         int64            `json:"bytesOut"`
	CpuPercent       float64          `json:"cpuPercent"`
	MemoryBytes      int64            `json:"memoryBytes"`
	MetricsScrapes   int64            `json:"metricsScrapes"`
	Latency          TestRunLatency   `json:"latency"`
}

// TestRunLatency mirrors the cumulative latency histogram of one agent.
type TestRunLatency struct {
	BoundsMs []float64 `json:"boundsMs"`
	Counts   []int64   `json:"counts"`
	Total    int64     `json:"total"`
	SumMs    float64   `json:"sumMs"`
}

// TestRunSample is one fleet time-series point (interval deltas, not cumulative).
type TestRunSample struct {
	TimestampMs      int64   `json:"timestampMs"`
	EgressRps        float64 `json:"egressRps"`
	P50Ms            float64 `json:"p50Ms"`
	P90Ms            float64 `json:"p90Ms"`
	P95Ms            float64 `json:"p95Ms"`
	P99Ms            float64 `json:"p99Ms"`
	Ok               int64   `json:"ok"`
	ClientErrors     int64   `json:"clientErrors"`
	ServerErrors     int64   `json:"serverErrors"`
	Timeouts         int64   `json:"timeouts"`
	Resets           int64   `json:"resets"`
	ConnectionErrors int64   `json:"connectionErrors"`
	ErrorRate        float64 `json:"errorRate"`
}
