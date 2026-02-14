package main

import (
	"fmt"
	"log"
	"os"

	"mallekoppie/ChaosGenerator/internal/client"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
)

func main() {
	// Initialize platform for logging
	config := platform.Config{}
	config.Component.ComponentName = "chaos-client"
	platform.SetPlatformConfiguration(config)

	// Get server address from environment or use default
	serverAddress := os.Getenv("CHAOS_MASTER_ADDRESS")
	if serverAddress == "" {
		serverAddress = "127.0.0.1:9002"
	}

	platform.Log.Info("Connecting to ChaosMaster server", zap.String("address", serverAddress))

	// Create client
	c, err := client.NewClient(serverAddress)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	platform.Log.Info("Connected successfully to ChaosMaster server")
	fmt.Println("Welcome to ChaosMaster Test Client!")
	fmt.Println("Server:", serverAddress)

	// Start interactive menu
	c.Start()
}
