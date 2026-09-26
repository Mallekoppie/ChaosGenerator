package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	pb "mallekoppie/ChaosGenerator/internal/contracts"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// heartbeatInterval must stay comfortably below the master's heartbeat
	// timeout so a couple of dropped beats are tolerated before the master
	// declares the connection dead.
	heartbeatInterval = 10 * time.Second
)

// Client represents an agent client that connects to the master
type Client struct {
	masterAddress string
	agentId       string
	handler       *CommandHandler

	// sendMu serialises writes to the stream. gRPC forbids concurrent SendMsg
	// calls and both the heartbeat and command loops write to the same stream.
	sendMu sync.Mutex
}

// NewClient creates a new agent client
func NewClient(masterAddress string, agentId string) *Client {
	return &Client{
		masterAddress: masterAddress,
		agentId:       agentId,
		handler:       NewCommandHandler(),
	}
}

// ConnectAndServe establishes the stream to the master and blocks until it
// ends, returning the reason. The caller is expected to reconnect.
func (c *Client) ConnectAndServe(ctx context.Context) error {
	log.Printf("Agent connecting to master at %s", c.masterAddress)

	conn, err := grpc.Dial(c.masterAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to dial master: %w", err)
	}
	defer conn.Close()

	client := pb.NewChaosMasterClient(conn)

	stream, err := client.ConnectAgent(ctx)
	if err != nil {
		return fmt.Errorf("failed to establish stream: %w", err)
	}

	// Send one heartbeat straight away so the master registers this connection
	// immediately instead of waiting for the first periodic beat.
	if err := c.send(stream, c.heartbeatMessage()); err != nil {
		return fmt.Errorf("failed to send handshake heartbeat: %w", err)
	}

	log.Printf("Agent %s connected to master", c.agentId)

	heartbeatCtx, cancelHeartbeats := context.WithCancel(ctx)
	defer cancelHeartbeats()
	go c.sendHeartbeats(heartbeatCtx, stream)

	return c.receiveCommands(ctx, stream)
}

// heartbeatMessage builds a heartbeat for this agent.
func (c *Client) heartbeatMessage() *pb.AgentMessage {
	return &pb.AgentMessage{
		AgentId: c.agentId,
		Payload: &pb.AgentMessage_Heartbeat{
			Heartbeat: &pb.AgentHeartbeat{
				Timestamp: time.Now().Unix(),
			},
		},
	}
}

// send writes a message to the stream, serialising concurrent writers.
func (c *Client) send(stream pb.ChaosMaster_ConnectAgentClient, msg *pb.AgentMessage) error {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()

	return stream.Send(msg)
}

// sendHeartbeats sends periodic heartbeats to the master
func (c *Client) sendHeartbeats(ctx context.Context, stream pb.ChaosMaster_ConnectAgentClient) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.send(stream, c.heartbeatMessage()); err != nil {
				log.Printf("Error sending heartbeat, dropping connection: %v", err)

				// Half-close the stream so the master stops waiting for us and the
				// receive loop unblocks; the runner then reconnects.
				_ = stream.CloseSend()
				return
			}
		}
	}
}

// receiveCommands listens for commands from the master and blocks until the
// stream ends.
func (c *Client) receiveCommands(ctx context.Context, stream pb.ChaosMaster_ConnectAgentClient) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		cmd, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("Master closed the connection")
				return nil
			}
			return err
		}

		log.Printf("Received command: %s", cmd.CommandId)

		response := c.handler.HandleCommand(cmd)

		msg := &pb.AgentMessage{
			AgentId: c.agentId,
			Payload: &pb.AgentMessage_Response{
				Response: response,
			},
		}

		if err := c.send(stream, msg); err != nil {
			return err
		}
	}
}
