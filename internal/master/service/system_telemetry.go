package service

import (
	"context"
	"runtime"
	"sync"
	"time"

	"mallekoppie/ChaosGenerator/internal/sysmetrics"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
)

// serverResourceSampleInterval is how often the master samples its own process.
// It matches the UI poll cadence so each reading covers a comparable window.
const serverResourceSampleInterval = 3 * time.Second

// ServerResourceUsage is the master process resource footprint plus a couple of
// runtime facts that help explain memory growth.
type ServerResourceUsage struct {
	CPUPercent    float64
	MemoryBytes   int64
	Goroutines    int32
	UptimeSeconds int64
}

var (
	serverStartedAt = time.Now()

	serverResourceMu      sync.RWMutex
	serverCPUPercent      float64
	serverMemoryBytes     int64
	serverResourceSampled bool

	// serverFallbackSampler is used when the background sampler was not started
	// (for example in unit tests) so the RPC still returns a reading.
	serverFallbackSampler = sysmetrics.NewProcessSampler()

	serverSamplerOnce sync.Once
)

// StartServerResourceSampler samples the master process on a fixed interval so
// every telemetry poll reads a value measured over a consistent window. It is a
// no-op after the first call.
func StartServerResourceSampler(ctx context.Context) {
	serverSamplerOnce.Do(func() {
		sampler := sysmetrics.NewProcessSampler()

		go func() {
			ticker := time.NewTicker(serverResourceSampleInterval)
			defer ticker.Stop()

			// Prime the first delta so the first tick reports a usable
			// percentage instead of the sampler's initial 0%.
			recordServerResource(sampler.Sample())

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					recordServerResource(sampler.Sample())
				}
			}
		}()

		platform.Log.Info("Server resource sampler started",
			zap.Duration("interval", serverResourceSampleInterval))
	})
}

func recordServerResource(cpuPercent float64, memoryBytes int64) {
	serverResourceMu.Lock()
	defer serverResourceMu.Unlock()

	serverCPUPercent = cpuPercent
	serverMemoryBytes = memoryBytes
	serverResourceSampled = true
}

// serverResourceUsage returns the latest master process reading. When the
// background sampler is not running it falls back to an on-demand sample.
func serverResourceUsage() ServerResourceUsage {
	serverResourceMu.RLock()
	usage := ServerResourceUsage{
		CPUPercent:  serverCPUPercent,
		MemoryBytes: serverMemoryBytes,
	}
	sampled := serverResourceSampled
	serverResourceMu.RUnlock()

	if !sampled {
		usage.CPUPercent, usage.MemoryBytes = serverFallbackSampler.Sample()
	}

	usage.Goroutines = int32(runtime.NumGoroutine())
	usage.UptimeSeconds = int64(time.Since(serverStartedAt).Seconds())

	return usage
}
