package agent

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ProcessSampler reports process-level CPU and memory usage for the agent.
//
// CPU is derived from the delta of cumulative Getrusage CPU time between
// samples, so the first Sample call after start reports 0% and every subsequent
// call reports the average utilisation over the elapsed interval. Memory is the
// process resident set size where the platform exposes it, falling back to the
// Go runtime's system allocation.
type ProcessSampler struct {
	mu       sync.Mutex
	lastCPU  time.Duration
	lastWall time.Time
}

// NewProcessSampler creates a sampler ready for its first reading.
func NewProcessSampler() *ProcessSampler {
	return &ProcessSampler{}
}

// Sample returns the CPU percentage consumed since the previous call and the
// current process memory footprint in bytes.
func (s *ProcessSampler) Sample() (cpuPercent float64, memoryBytes int64) {
	now := time.Now()
	cpu := cumulativeCPUTime()

	s.mu.Lock()
	lastCPU, lastWall := s.lastCPU, s.lastWall
	s.lastCPU, s.lastWall = cpu, now
	s.mu.Unlock()

	if !lastWall.IsZero() {
		wall := now.Sub(lastWall)
		if wall > 0 && cpu >= lastCPU {
			cpuPercent = float64(cpu-lastCPU) / float64(wall) * 100
		}
	}

	return cpuPercent, processMemoryBytes()
}

// cumulativeCPUTime returns the total user+system CPU time consumed by the
// process so far.
func cumulativeCPUTime() time.Duration {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0
	}
	return time.Duration(ru.Utime.Nano() + ru.Stime.Nano())
}

// processMemoryBytes returns the resident set size of the process, falling back
// to the Go runtime's system allocation when /proc is unavailable.
func processMemoryBytes() int64 {
	if rss, ok := readVmRSS(); ok {
		return rss
	}

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return int64(stats.Sys)
}

// readVmRSS reads VmRSS from /proc/self/status, the closest portable
// approximation of the container's resident memory on Linux.
func readVmRSS() (int64, bool) {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, false
	}

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0, false
		}

		kilobytes, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return 0, false
		}

		return kilobytes * 1024, true
	}

	return 0, false
}
