package service

import (
	"io"
	"log"
	"sync"

	pb "mallekoppie/ChaosGenerator/internal/contracts"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
)

// AgentConnection represents an active agent connection
type AgentConnection struct {
	AgentId string
	Stream  pb.ChaosMaster_ConnectAgentServer
}

// ConnectionManager manages all active agent connections
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]*AgentConnection
}

var (
	connManager *ConnectionManager
	once        sync.Once
)

// GetConnectionManager returns the singleton connection manager
func GetConnectionManager() *ConnectionManager {
	once.Do(func() {
		connManager = &ConnectionManager{
			connections: make(map[string]*AgentConnection),
		}
	})
	return connManager
}

// AddConnection adds a new agent connection
func (cm *ConnectionManager) AddConnection(agentId string, stream pb.ChaosMaster_ConnectAgentServer) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.connections[agentId] = &AgentConnection{
		AgentId: agentId,
		Stream:  stream,
	}
	platform.Log.Info("Agent connected", zap.String("agentId", agentId))
}

// RemoveConnection removes an agent connection
func (cm *ConnectionManager) RemoveConnection(agentId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.connections, agentId)
	platform.Log.Info("Agent disconnected", zap.String("agentId", agentId))
}

// GetConnection gets a connection by agent ID
func (cm *ConnectionManager) GetConnection(agentId string) (*AgentConnection, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, ok := cm.connections[agentId]
	return conn, ok
}

// GetAllConnections returns all active connections
func (cm *ConnectionManager) GetAllConnections() []*AgentConnection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conns := make([]*AgentConnection, 0, len(cm.connections))
	for _, conn := range cm.connections {
		conns = append(conns, conn)
	}
	return conns
}

// SendCommand sends a command to a specific agent
func (cm *ConnectionManager) SendCommand(agentId string, command *pb.AgentCommand) (*pb.CommandResponse, error) {
	conn, ok := cm.GetConnection(agentId)
	if !ok {
		platform.Log.Warn("Agent not connected", zap.String("agentId", agentId))
		return nil, ErrAgentNotConnected
	}

	// Send command
	if err := conn.Stream.Send(command); err != nil {
		platform.Log.Error("Error sending command to agent", zap.String("agentId", agentId), zap.Error(err))
		cm.RemoveConnection(agentId)
		return nil, err
	}

	platform.Log.Info("Command sent to agent", zap.String("agentId", agentId), zap.String("commandId", command.CommandId))
	return nil, nil
}

// ConnectAgent handles the bidirectional streaming connection from an agent
func (s *ChaosMasterServer) ConnectAgent(stream pb.ChaosMaster_ConnectAgentServer) error {
	platform.Log.Info("New agent connection established")

	var agentId string
	cm := GetConnectionManager()

	// Receive first message to get agent ID
	msg, err := stream.Recv()
	if err != nil {
		platform.Log.Error("Error receiving initial message from agent", zap.Error(err))
		return err
	}

	agentId = msg.AgentId
	if agentId == "" {
		platform.Log.Error("Agent ID not provided in initial message")
		return ErrInvalidAgentId
	}

	// Register the connection
	cm.AddConnection(agentId, stream)
	defer cm.RemoveConnection(agentId)

	// Handle incoming messages from agent
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			platform.Log.Info("Agent closed connection", zap.String("agentId", agentId))
			return nil
		}
		if err != nil {
			platform.Log.Error("Error receiving message from agent", zap.String("agentId", agentId), zap.Error(err))
			return err
		}

		// Handle different message types
		switch payload := msg.Payload.(type) {
		case *pb.AgentMessage_Heartbeat:
			platform.Log.Debug("Received heartbeat from agent", zap.String("agentId", agentId))
			// Heartbeat received - connection is alive

		case *pb.AgentMessage_Response:
			platform.Log.Info("Received command response from agent",
				zap.String("agentId", agentId),
				zap.String("commandId", payload.Response.CommandId),
				zap.Bool("success", payload.Response.Success))

			// Handle command response
			// In a real implementation, you might want to store these responses
			// or notify waiting clients
			if !payload.Response.Success {
				log.Printf("Command failed: %s", payload.Response.Error)
			}

		default:
			platform.Log.Warn("Unknown message type from agent", zap.String("agentId", agentId))
		}
	}
}
