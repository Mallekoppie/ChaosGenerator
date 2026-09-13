package targethttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/http2"
)

type Server struct {
	config        Config
	httpServer    *http.Server
	metricsServer *http.Server
}

func NewServer(config Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Start() error {
	// Setup main HTTP server
	mux := http.NewServeMux()

	// Register use case handlers
	mux.HandleFunc("/health", UC1Handler)
	mux.HandleFunc("/api/v1/status", UC2Handler)
	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Check if it's a list or single query
			if r.URL.Query().Has("page") || r.URL.Query().Has("limit") {
				UC5Handler(w, r)
			} else {
				UC3Handler(w, r)
			}
		} else if r.Method == "POST" {
			UC4Handler(w, r)
		}
	})
	mux.HandleFunc("/api/v1/users/bulk", UC6Handler)
	mux.HandleFunc("/api/v1/reports/generate", UC7Handler)
	mux.HandleFunc("/api/v1/documents/upload", UC8Handler)
	mux.HandleFunc("/api/v1/transactions/export", UC9Handler)
	mux.HandleFunc("/api/v1/data/process", UC10Handler)
	mux.HandleFunc("/api/v1/analytics/stream", UC11Handler)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Configure HTTP/2 if specified
	if s.config.Protocol == "http2" {
		// For HTTP/2, we need TLS
		s.httpServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{"h2", "http/1.1"},
		}
		if err := http2.ConfigureServer(s.httpServer, &http2.Server{}); err != nil {
			return fmt.Errorf("failed to configure HTTP/2: %w", err)
		}
	} else if s.config.Protocol == "http1" && s.config.HasTLS() {
		// For HTTP/1.1 with TLS
		s.httpServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	// Setup metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	s.metricsServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.MetricsPort),
		Handler: metricsMux,
	}

	// Start metrics server
	go func() {
		log.Printf("Starting metrics server on port %d", s.config.MetricsPort)
		if err := s.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Start main server
	log.Printf("Starting %s server on port %d", s.config.Protocol, s.config.Port)

	// Determine if TLS should be used
	useTLS := s.config.HasTLS() || s.config.Protocol == "http2"

	if useTLS {
		if !s.config.HasTLS() {
			// HTTP/2 requires TLS but no certificates provided
			return fmt.Errorf("HTTP/2 requires TLS certificates. Please set CHAOS_TARGET_CERT_FILE and CHAOS_TARGET_KEY_FILE environment variables")
		}
		log.Printf("Starting with TLS using cert: %s, key: %s", s.config.CertFile, s.config.KeyFile)
		return s.httpServer.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
	}

	log.Printf("Starting without TLS (plain HTTP)")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.metricsServer.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down metrics server: %v", err)
	}
	return s.httpServer.Shutdown(ctx)
}
