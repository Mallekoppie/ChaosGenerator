package agent

import (
	"context"
	"log"
	"os"
	"strconv"

	pb "mallekoppie/ChaosGenerator/internal/contracts"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RegisterWithMaster registers the agent with the ChaosMaster server
func RegisterWithMaster(masterAddress string, port string, metricsPort string, version string) (string, error) {
	log.Printf("Registering with ChaosMaster at %s", masterAddress)

	// Connect to master server (using insecure for now, matching master config)
	conn, err := grpc.Dial(masterAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to master: %v", err)
		return "", err
	}
	defer conn.Close()

	client := pb.NewChaosMasterClient(conn)

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Failed to get hostname, using 'unknown': %v", err)
		hostname = "unknown"
	}

	// Parse port as int32
	portInt, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		log.Printf("Failed to parse port: %v", err)
		return "", err
	}

	metricsPortInt, err := strconv.ParseInt(metricsPort, 10, 32)
	if err != nil {
		log.Printf("Failed to parse metrics port: %v", err)
		return "", err
	}

	// Register with master
	req := &pb.RegisterAgentRequest{
		Hostname:    hostname,
		Port:        int32(portInt),
		MetricsPort: int32(metricsPortInt),
		Version:     version,
	}

	resp, err := client.RegisterAgent(context.Background(), req)
	if err != nil {
		log.Printf("Failed to register with master: %v", err)
		return "", err
	}

	if resp.Success {
		log.Printf("Successfully registered with master. Agent ID: %s", resp.AgentId)
		return resp.AgentId, nil
	} else {
		log.Printf("Registration failed: %s", resp.Message)
		return "", nil
	}
}
