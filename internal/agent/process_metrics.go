package agent

import (
	"time"

	"mallekoppie/ChaosGenerator/internal/sysmetrics"
)

// ProcessSampler reports process-level CPU and memory usage for the agent. It
// aliases the shared sampler so the agent and the master measure resource usage
// identically.
type ProcessSampler = sysmetrics.ProcessSampler

// NewProcessSampler creates a sampler ready for its first reading.
func NewProcessSampler() *ProcessSampler {
	return sysmetrics.NewProcessSampler()
}

// cumulativeCPUTime returns the total user+system CPU time consumed by the
// process so far.
func cumulativeCPUTime() time.Duration {
	return sysmetrics.CumulativeCPUTime()
}
