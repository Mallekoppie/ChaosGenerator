package targetgrpc

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"mallekoppie/ChaosGenerator/internal/contracts"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	config        Config
	grpcServer    *grpc.Server
	metricsServer *http.Server
}

func NewServer(config Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Start() error {
	// Start metrics server
	go s.startMetricsServer()

	// Start gRPC server
	return s.startGRPCServer()
}

func (s *Server) startGRPCServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create gRPC server with options
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(256 * 1024 * 1024), // 256MB max receive
		grpc.MaxSendMsgSize(256 * 1024 * 1024), // 256MB max send
	}

	// Add TLS credentials if configured
	if s.config.HasTLS() {
		creds, err := credentials.NewServerTLSFromFile(s.config.CertFile, s.config.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
		log.Printf("TLS enabled with cert: %s", s.config.CertFile)
	} else {
		log.Println("TLS not configured, running without encryption")
	}

	s.grpcServer = grpc.NewServer(opts...)

	// Register service
	service := NewChaosTargetService()
	contracts.RegisterChaosTargetServer(s.grpcServer, service)

	// Register reflection service for grpcurl
	reflection.Register(s.grpcServer)

	log.Printf("gRPC server listening on port %d", s.config.Port)

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) startMetricsServer() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	s.metricsServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.MetricsPort),
		Handler: mux,
	}

	log.Printf("Metrics server listening on port %d", s.config.MetricsPort)

	if err := s.metricsServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("Metrics server error: %v", err)
	}
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		log.Println("Shutting down gRPC server...")
		s.grpcServer.GracefulStop()
	}

	if s.metricsServer != nil {
		log.Println("Shutting down metrics server...")
		s.metricsServer.Close()
	}
}
