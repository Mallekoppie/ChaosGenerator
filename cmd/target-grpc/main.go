package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	targetgrpc "mallekoppie/ChaosGenerator/internal/target-grpc"
)

func main() {
	log.Println("Starting Chaos Target gRPC Service")

	// Load configuration
	config := targetgrpc.LoadConfig()
	log.Printf("Configuration: Port=%d, MetricsPort=%d",
		config.Port, config.MetricsPort)

	// Create server
	server := targetgrpc.NewServer(config)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		server.Stop()
		os.Exit(0)
	}()

	// Start server (blocks until shutdown)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
