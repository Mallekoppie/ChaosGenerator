package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsScrapeCount atomic.Int64
	metricsLastScrape  atomic.Int64
)

// MetricsScrapeStats returns the number of Prometheus scrapes of the agent's
// /metrics endpoint and the Unix timestamp of the most recent one. It is zero
// when the endpoint has never been scraped, which is how the UI can tell that
// an agent is running without Prometheus.
func MetricsScrapeStats() (count int64, lastUnix int64) {
	return metricsScrapeCount.Load(), metricsLastScrape.Load()
}

// StartMetricsServer starts an HTTP server to expose Prometheus metrics
func StartMetricsServer(ctx context.Context, port string) error {
	mux := http.NewServeMux()

	// Count scrapes so the agent can report Prometheus scraper health to the
	// master. The handler is built once and reused for every request.
	promHandler := promhttp.Handler()
	mux.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsScrapeCount.Add(1)
		metricsLastScrape.Store(time.Now().Unix())
		promHandler.ServeHTTP(w, r)
	}))

	// Add a health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in a goroutine
	go func() {
		log.Printf("Starting metrics server on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		log.Println("Shutting down metrics server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down metrics server: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Verify the server is accessible
	healthURL := fmt.Sprintf("http://localhost:%s/health", port)
	resp, err := http.Get(healthURL)
	if err == nil {
		resp.Body.Close()
		log.Printf("Metrics server is accessible at http://localhost:%s/metrics", port)
	}

	return nil
}
