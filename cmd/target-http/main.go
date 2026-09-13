package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	targethttp "mallekoppie/ChaosGenerator/internal/target-http"
)

func main() {
	log.Println("Starting Chaos Target HTTP Service")

	// Load configuration
	config := targethttp.LoadConfig()
	log.Printf("Configuration: Protocol=%s, Port=%d, MetricsPort=%d",
		config.Protocol, config.Port, config.MetricsPort)

	// Create and start server
	server := targethttp.NewServer(config)

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	// Start server (blocks until shutdown)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
