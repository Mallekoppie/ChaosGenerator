package agent

import (
	"context"
	"io"
	"log"
	"time"

	pb "mallekoppie/ChaosGenerator/internal/contracts"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents an agent client that connects to the master
type Client struct {
	masterAddress string
	agentId       string
	handler       *CommandHandler
	stream        pb.ChaosMaster_ConnectAgentClient
	conn          *grpc.ClientConn
}

// NewClient creates a new agent client
func NewClient(masterAddress string, agentId string) *Client {
	return &Client{
		masterAddress: masterAddress,
		agentId:       agentId,
		handler:       NewCommandHandler(),
	}
}

// Connect establishes connection to master and starts listening for commands
func (c *Client) Connect(ctx context.Context) error {
	log.Printf("Agent connecting to master at %s", c.masterAddress)

	// Connect to master server
	conn, err := grpc.Dial(c.masterAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to master: %v", err)
		return err
	}
	c.conn = conn

	client := pb.NewChaosMasterClient(conn)

	// Establish bidirectional stream
	stream, err := client.ConnectAgent(ctx)
	if err != nil {
		log.Printf("Failed to establish stream: %v", err)
		conn.Close()
		return err
	}
	c.stream = stream

	log.Printf("Agent %s connected to master", c.agentId)

	// Start sending heartbeats
	go c.sendHeartbeats(ctx)

	// Start receiving commands
	go c.receiveCommands(ctx)

	return nil
}

// sendHeartbeats sends periodic heartbeats to the master
func (c *Client) sendHeartbeats(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msg := &pb.AgentMessage{
				AgentId: c.agentId,
				Payload: &pb.AgentMessage_Heartbeat{
					Heartbeat: &pb.AgentHeartbeat{
						Timestamp: time.Now().Unix(),
					},
				},
			}

			if err := c.stream.Send(msg); err != nil {
				log.Printf("Error sending heartbeat: %v", err)
				return
			}
		}
	}
}

// receiveCommands listens for commands from the master
func (c *Client) receiveCommands(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			cmd, err := c.stream.Recv()
			if err == io.EOF {
				log.Println("Master closed the connection")
				return
			}
			if err != nil {
				log.Printf("Error receiving command: %v", err)
				return
			}

			log.Printf("Received command: %s", cmd.CommandId)

			// Process command
			response := c.handler.HandleCommand(cmd)

			// Send response back to master
			msg := &pb.AgentMessage{
				AgentId: c.agentId,
				Payload: &pb.AgentMessage_Response{
					Response: response,
				},
			}

			if err := c.stream.Send(msg); err != nil {
				log.Printf("Error sending response: %v", err)
				return
			}
		}
	}
}

// Close closes the connection to the master
func (c *Client) Close() error {
	if c.stream != nil {
		if err := c.stream.CloseSend(); err != nil {
			log.Printf("Error closing stream: %v", err)
		}
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
