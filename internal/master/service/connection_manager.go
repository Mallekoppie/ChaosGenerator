package service

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	pb "mallekoppie/ChaosGenerator/internal/contracts"
	"mallekoppie/ChaosGenerator/internal/master/repositories"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
)

const (
	// agentHeartbeatTimeout is how long the master waits for a heartbeat before
	// treating a connection as dead. Agents heartbeat every 10s, so this
	// tolerates two missed beats before reaping a connection whose socket never
	// closed cleanly (for example a pod deleted mid-flight).
	agentHeartbeatTimeout = 30 * time.Second
	// connectionMonitorInterval is how often stale connections are swept.
	connectionMonitorInterval = 5 * time.Second
	// agentCommandTimeout bounds how long a command write may block so a single
	// wedged agent cannot stall a fleet-wide test start.
	agentCommandTimeout = 5 * time.Second
)

// AgentConnection represents an active agent connection
type AgentConnection struct {
	AgentId       string
	Stream        pb.ChaosMaster_ConnectAgentServer
	ConnectedAt   time.Time
	LastHeartbeat time.Time

	// sendMu serialises writes to the stream. gRPC streams are not safe for
	// concurrent SendMsg calls and several tests can be started at once.
	sendMu sync.Mutex
	// cancel terminates the ConnectAgent handler that owns this connection.
	cancel context.CancelFunc
}

// send writes a command to the agent, serialising concurrent writers. The write
// runs in its own goroutine so a wedged peer cannot block the caller forever.
func (c *AgentConnection) send(command *pb.AgentCommand) error {
	result := make(chan error, 1)

	go func() {
		c.sendMu.Lock()
		defer c.sendMu.Unlock()
		result <- c.Stream.Send(command)
	}()

	select {
	case err := <-result:
		return err
	case <-time.After(agentCommandTimeout):
		return fmt.Errorf("timed out after %s sending command to agent %s", agentCommandTimeout, c.AgentId)
	}
}

// ConnectionManager manages all active agent connections
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]*AgentConnection
}

var (
	connManager *ConnectionManager
	once        sync.Once
	monitorOnce sync.Once
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

// StartConnectionMonitor reaps connections that have stopped sending heartbeats
// so the online agent list reflects reality even when a socket dies silently.
func StartConnectionMonitor(ctx context.Context) {
	cm := GetConnectionManager()

	monitorOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(connectionMonitorInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if reaped := cm.ReapStaleConnections(agentHeartbeatTimeout); reaped > 0 {
						platform.Log.Warn("Reaped unresponsive agent connections", zap.Int("count", reaped))
					}
				}
			}
		}()

		platform.Log.Info("Agent connection monitor started",
			zap.Duration("heartbeatTimeout", agentHeartbeatTimeout),
			zap.Duration("sweepInterval", connectionMonitorInterval))
	})
}

// AddConnection registers a new agent connection and returns it. Any previous
// connection for the same agent ID (a fast reconnect) is terminated.
func (cm *ConnectionManager) AddConnection(agentId string, stream pb.ChaosMaster_ConnectAgentServer, cancel context.CancelFunc) *AgentConnection {
	now := time.Now()
	conn := &AgentConnection{
		AgentId:       agentId,
		Stream:        stream,
		ConnectedAt:   now,
		LastHeartbeat: now,
		cancel:        cancel,
	}

	var replaced *AgentConnection

	cm.mu.Lock()
	if previous, ok := cm.connections[agentId]; ok && previous != conn {
		replaced = previous
	}
	cm.connections[agentId] = conn
	cm.mu.Unlock()

	if replaced != nil && replaced.cancel != nil {
		replaced.cancel()
	}

	platform.Log.Info("Agent connected", zap.String("agentId", agentId))
	return conn
}

// RemoveConnection removes an agent connection and deletes it from the database.
// It is a no-op when the connection was already reaped or replaced, so a stale
// handler returning late never clobbers a newer stream.
func (cm *ConnectionManager) RemoveConnection(conn *AgentConnection) {
	if conn == nil {
		return
	}

	cm.mu.Lock()
	current, ok := cm.connections[conn.AgentId]
	if !ok || current != conn {
		cm.mu.Unlock()
		return
	}
	delete(cm.connections, conn.AgentId)
	cm.mu.Unlock()

	platform.Log.Info("Agent disconnected",
		zap.String("agentId", conn.AgentId),
		zap.Duration("connectedFor", time.Since(conn.ConnectedAt)))

	// The agent registry mirrors the live connections: a disconnected agent is
	// removed from the database and re-registers when it reconnects.
	if err := repositories.DeleteAgent(conn.AgentId); err != nil {
		platform.Log.Error("Error deleting agent from database on disconnect",
			zap.String("agentId", conn.AgentId),
			zap.Error(err))
	}
}

// TouchHeartbeat records that an agent is still alive.
func (cm *ConnectionManager) TouchHeartbeat(agentId string, at time.Time) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, ok := cm.connections[agentId]; ok {
		conn.LastHeartbeat = at
	}
}

// ReapStaleConnections drops connections that have not sent a heartbeat within
// timeout and returns how many were reaped. Returning from the owning handler
// closes the stream, so the agent notices and reconnects.
func (cm *ConnectionManager) ReapStaleConnections(timeout time.Duration) int {
	now := time.Now()

	cm.mu.Lock()
	stale := make([]*AgentConnection, 0)
	for agentId, conn := range cm.connections {
		if now.Sub(conn.LastHeartbeat) > timeout {
			delete(cm.connections, agentId)
			stale = append(stale, conn)
		}
	}
	cm.mu.Unlock()

	for _, conn := range stale {
		platform.Log.Warn("Agent stopped heartbeating, dropping connection",
			zap.String("agentId", conn.AgentId),
			zap.Duration("silentFor", now.Sub(conn.LastHeartbeat)))

		if conn.cancel != nil {
			conn.cancel()
		}

		if err := repositories.DeleteAgent(conn.AgentId); err != nil {
			platform.Log.Error("Error deleting reaped agent from database",
				zap.String("agentId", conn.AgentId),
				zap.Error(err))
		}
	}

	return len(stale)
}

// IsAgentConnected reports whether a live connection exists for the agent.
func (cm *ConnectionManager) IsAgentConnected(agentId string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	_, ok := cm.connections[agentId]
	return ok
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

// GetAllAgentIds returns all active agent IDs
func (cm *ConnectionManager) GetAllAgentIds() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ids := make([]string, 0, len(cm.connections))
	for agentId := range cm.connections {
		ids = append(ids, agentId)
	}
	return ids
}

// CloseAll terminates every active connection. Used when the registry is
// cleared, so connected agents reconnect and register again.
func (cm *ConnectionManager) CloseAll() {
	cm.mu.Lock()
	conns := make([]*AgentConnection, 0, len(cm.connections))
	for agentId, conn := range cm.connections {
		delete(cm.connections, agentId)
		conns = append(conns, conn)
	}
	cm.mu.Unlock()

	for _, conn := range conns {
		if conn.cancel != nil {
			conn.cancel()
		}

		if err := repositories.DeleteAgent(conn.AgentId); err != nil {
			platform.Log.Error("Error deleting agent from database while closing all",
				zap.String("agentId", conn.AgentId),
				zap.Error(err))
		}
	}
}

// SendCommand sends a command to a specific agent
func (cm *ConnectionManager) SendCommand(agentId string, command *pb.AgentCommand) error {
	conn, ok := cm.GetConnection(agentId)
	if !ok {
		platform.Log.Warn("Agent not connected", zap.String("agentId", agentId))
		return ErrAgentNotConnected
	}

	if err := conn.send(command); err != nil {
		platform.Log.Error("Error sending command to agent", zap.String("agentId", agentId), zap.Error(err))

		// Tear the stream down so the handler removes the dead agent and it can
		// reconnect cleanly on a new stream.
		if conn.cancel != nil {
			conn.cancel()
		}

		return err
	}

	platform.Log.Info("Command sent to agent", zap.String("agentId", agentId), zap.String("commandId", command.CommandId))
	return nil
}

// ConnectAgent handles the bidirectional streaming connection from an agent
func (s *ChaosMasterServer) ConnectAgent(stream pb.ChaosMaster_ConnectAgentServer) error {
	platform.Log.Info("New agent connection established")

	cm := GetConnectionManager()

	// The first message carries the agent ID. Agents send a heartbeat immediately
	// after opening the stream, so the connection is registered without waiting
	// for the first periodic beat.
	msg, err := stream.Recv()
	if err != nil {
		platform.Log.Error("Error receiving initial message from agent", zap.Error(err))
		return err
	}

	agentId := msg.AgentId
	if agentId == "" {
		platform.Log.Error("Agent ID not provided in initial message")
		return ErrInvalidAgentId
	}

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	conn := cm.AddConnection(agentId, stream, cancel)
	defer cm.RemoveConnection(conn)

	handleAgentMessage(agentId, msg, cm)

	// Recv blocks until a message or a stream error arrives, so read it in a
	// goroutine and let the connection monitor cancel ctx to unwind a wedged
	// connection.
	messages := make(chan *pb.AgentMessage)
	recvErr := make(chan error, 1)

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				select {
				case recvErr <- err:
				case <-ctx.Done():
				}
				return
			}

			select {
			case messages <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			platform.Log.Info("Agent connection closed by master", zap.String("agentId", agentId))
			return nil

		case err := <-recvErr:
			if err == io.EOF {
				platform.Log.Info("Agent closed connection", zap.String("agentId", agentId))
				return nil
			}
			if ctx.Err() != nil {
				return nil
			}
			platform.Log.Error("Error receiving message from agent", zap.String("agentId", agentId), zap.Error(err))
			return err

		case msg := <-messages:
			handleAgentMessage(agentId, msg, cm)
		}
	}
}

// handleAgentMessage processes a single message received from an agent.
func handleAgentMessage(agentId string, msg *pb.AgentMessage, cm *ConnectionManager) {
	switch payload := msg.Payload.(type) {
	case *pb.AgentMessage_Heartbeat:
		// Liveness is tracked from the master's clock, not the agent's, so a
		// skewed agent clock cannot mask a dead connection.
		cm.TouchHeartbeat(agentId, time.Now())

	case *pb.AgentMessage_Response:
		if payload.Response == nil {
			return
		}

		platform.Log.Info("Received command response from agent",
			zap.String("agentId", agentId),
			zap.String("commandId", payload.Response.CommandId),
			zap.Bool("success", payload.Response.Success))

		if !payload.Response.Success {
			platform.Log.Warn("Agent command failed",
				zap.String("agentId", agentId),
				zap.String("commandId", payload.Response.CommandId),
				zap.String("error", payload.Response.Error))
		}

	case *pb.AgentMessage_MetricsReport:
		if payload.MetricsReport == nil {
			return
		}

		// Telemetry ingest is intentionally quiet: it arrives about once a
		// second per running test and would otherwise dominate the logs.
		GetTestMetricsStore().Ingest(agentId, payload.MetricsReport)

	default:
		platform.Log.Warn("Unknown message type from agent", zap.String("agentId", agentId))
	}
}
