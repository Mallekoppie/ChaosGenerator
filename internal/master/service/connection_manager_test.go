package service

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	pb "mallekoppie/ChaosGenerator/internal/contracts"
	"mallekoppie/ChaosGenerator/internal/master/logic"
	"mallekoppie/ChaosGenerator/internal/master/models"

	"github.com/Mallekoppie/goslow/platform"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestMain(m *testing.M) {
	dbFile := filepath.Join(os.TempDir(), fmt.Sprintf("chaos_service_%d.db", time.Now().UnixNano()))

	config := platform.Config{}
	config.Component.ComponentName = "chaos-service-tests"
	config.Log.Level = "error"
	config.Database.BoltDB.Enabled = true
	config.Database.BoltDB.FileName = dbFile

	platform.SetPlatformConfiguration(config)

	if err := platform.SetupBoltDB(); err != nil {
		fmt.Println("Unable to set up BoltDB for tests:", err)
		os.Exit(1)
	}

	code := m.Run()

	_ = platform.Database.BoltDb.Close()
	os.Remove(dbFile)
	os.Remove(dbFile + ".lock")

	os.Exit(code)
}

// fakeAgent plays the client half of the ConnectAgent stream the way the real
// agent binary does: an immediate handshake heartbeat, periodic heartbeats, and
// a response for every command.
type fakeAgent struct {
	id string

	mu       sync.Mutex
	sendMu   sync.Mutex
	stream   pb.ChaosMaster_ConnectAgentClient
	commands []*pb.AgentCommand

	cancel        context.CancelFunc
	stopHeartbeat context.CancelFunc
}

func (a *fakeAgent) send(msg *pb.AgentMessage) error {
	a.sendMu.Lock()
	defer a.sendMu.Unlock()

	return a.stream.Send(msg)
}

func (a *fakeAgent) heartbeat() error {
	return a.send(&pb.AgentMessage{
		AgentId: a.id,
		Payload: &pb.AgentMessage_Heartbeat{
			Heartbeat: &pb.AgentHeartbeat{Timestamp: time.Now().Unix()},
		},
	})
}

func (a *fakeAgent) startHeartbeats(ctx context.Context, interval time.Duration) {
	hbCtx, cancel := context.WithCancel(ctx)
	a.stopHeartbeat = cancel

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-hbCtx.Done():
				return
			case <-ticker.C:
				if err := a.heartbeat(); err != nil {
					return
				}
			}
		}
	}()
}

func (a *fakeAgent) receivedStartTests() int {
	a.mu.Lock()
	defer a.mu.Unlock()

	count := 0
	for _, cmd := range a.commands {
		if cmd.GetStartTest() != nil {
			count++
		}
	}
	return count
}

func (a *fakeAgent) close() {
	a.goSilent()
	if a.cancel != nil {
		a.cancel()
	}
}

// goSilent stops heartbeating but leaves the stream open, which is what a pod
// that is killed without closing its socket looks like to the master.
func (a *fakeAgent) goSilent() {
	if a.stopHeartbeat != nil {
		a.stopHeartbeat()
	}
}

func startFakeAgent(t *testing.T, ctx context.Context, client pb.ChaosMasterClient, host string, heartbeat time.Duration) *fakeAgent {
	t.Helper()

	reg, err := client.RegisterAgent(ctx, &pb.RegisterAgentRequest{
		Hostname:    host,
		Port:        9001,
		MetricsPort: 9091,
		Version:     "test",
	})
	if err != nil {
		t.Fatalf("RegisterAgent(%s) failed: %v", host, err)
	}
	if !reg.Success {
		t.Fatalf("RegisterAgent(%s) rejected: %s", host, reg.Message)
	}

	streamCtx, cancel := context.WithCancel(ctx)
	stream, err := client.ConnectAgent(streamCtx)
	if err != nil {
		cancel()
		t.Fatalf("ConnectAgent(%s) failed: %v", host, err)
	}

	agent := &fakeAgent{id: reg.AgentId, stream: stream, cancel: cancel}

	if err := agent.heartbeat(); err != nil {
		cancel()
		t.Fatalf("handshake heartbeat for %s failed: %v", host, err)
	}

	agent.startHeartbeats(streamCtx, heartbeat)

	go func() {
		for {
			cmd, err := stream.Recv()
			if err != nil {
				return
			}

			agent.mu.Lock()
			agent.commands = append(agent.commands, cmd)
			agent.mu.Unlock()

			_ = agent.send(&pb.AgentMessage{
				AgentId: agent.id,
				Payload: &pb.AgentMessage_Response{
					Response: &pb.CommandResponse{CommandId: cmd.CommandId, Success: true},
				},
			})
		}
	}()

	return agent
}

func startTestServer(t *testing.T) (pb.ChaosMasterClient, *grpc.Server) {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterChaosMasterServer(server, &ChaosMasterServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("test gRPC server stopped: %v", err)
		}
	}()

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial test server: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
		server.Stop()
	})

	return pb.NewChaosMasterClient(conn), server
}

func registerTestTarget(t *testing.T) {
	t.Helper()

	_, err := logic.RegisterTarget(models.Target{
		ID:       "test-target",
		Name:     "Test Target",
		Address:  "127.0.0.1:1",
		Protocol: "http",
	})
	if err != nil {
		t.Fatalf("failed to register test target: %v", err)
	}
}

func waitFor(t *testing.T, timeout time.Duration, what string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s", what)
}

func connectedAgents(t *testing.T, client pb.ChaosMasterClient) []*pb.Agent {
	t.Helper()

	resp, err := client.GetAllAgents(context.Background(), &pb.GetAllAgentsRequest{})
	if err != nil {
		t.Fatalf("GetAllAgents failed: %v", err)
	}

	return resp.Agents
}

func resetState(t *testing.T, agents ...*fakeAgent) {
	t.Helper()

	for _, agent := range agents {
		agent.close()
	}

	GetConnectionManager().CloseAll()

	if _, err := logic.ClearAllAgents(); err != nil {
		t.Logf("failed to clear agents during cleanup: %v", err)
	}
}

// TestAgentsAreListedLiveAndTestsReachEveryAgent covers the reported bugs: the
// agent list must reflect live connections, and a test must be delivered to
// every connected agent rather than a single one.
func TestAgentsAreListedLiveAndTestsReachEveryAgent(t *testing.T) {
	registerTestTarget(t)

	client, _ := startTestServer(t)
	ctx := context.Background()

	agentA := startFakeAgent(t, ctx, client, "agent-a", 20*time.Millisecond)
	agentB := startFakeAgent(t, ctx, client, "agent-b", 20*time.Millisecond)
	defer resetState(t, agentA, agentB)

	waitFor(t, 5*time.Second, "both agents to be listed", func() bool {
		return len(connectedAgents(t, client)) == 2
	})

	for _, agent := range connectedAgents(t, client) {
		if agent.Status != "online" {
			t.Errorf("expected agent %s to be online, got %q", agent.Id, agent.Status)
		}
	}

	resp, err := client.StartTestExecution(ctx, &pb.StartTestExecutionRequest{
		UseCaseId:              "uc1",
		TargetId:               "test-target",
		SimulatedUsersPerAgent: 5,
		AgentSelection:         "all",
	})
	if err != nil {
		t.Fatalf("StartTestExecution failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("StartTestExecution was rejected: %s", resp.Message)
	}

	waitFor(t, 5*time.Second, "both agents to receive the start command", func() bool {
		return agentA.receivedStartTests() == 1 && agentB.receivedStartTests() == 1
	})

	running, err := client.GetRunningTests(ctx, &pb.GetRunningTestsRequest{})
	if err != nil {
		t.Fatalf("GetRunningTests failed: %v", err)
	}

	var tracked int32
	for _, test := range running.Tests {
		if test.TestExecutionId == resp.TestExecutionId {
			tracked = test.NumberOfAgents
		}
	}
	if tracked != 2 {
		t.Errorf("expected the running test to track 2 agents, got %d", tracked)
	}
}

// TestStaleAgentIsReapedAndSkippedByTests covers the "deleted pod still shows
// online" bug: once an agent stops heartbeating it is dropped from the list and
// excluded from subsequent tests.
func TestStaleAgentIsReapedAndSkippedByTests(t *testing.T) {
	registerTestTarget(t)

	client, _ := startTestServer(t)
	ctx := context.Background()

	live := startFakeAgent(t, ctx, client, "agent-live", 20*time.Millisecond)
	dying := startFakeAgent(t, ctx, client, "agent-dying", 20*time.Millisecond)
	defer resetState(t, live, dying)

	waitFor(t, 5*time.Second, "both agents to be listed", func() bool {
		return len(connectedAgents(t, client)) == 2
	})

	// Simulate the pod disappearing: heartbeats stop but the stream is left open.
	dying.goSilent()

	// Let the dying agent go silent while the live one keeps beating, then reap
	// anything older than a threshold the live agent cannot reach.
	time.Sleep(500 * time.Millisecond)

	reaped := GetConnectionManager().ReapStaleConnections(250 * time.Millisecond)
	if reaped != 1 {
		t.Fatalf("expected 1 stale connection to be reaped, got %d", reaped)
	}

	waitFor(t, 5*time.Second, "the stale agent to leave the list", func() bool {
		agents := connectedAgents(t, client)
		return len(agents) == 1 && agents[0].Id == live.id
	})

	// The registry is the live connection set, so the row is gone too.
	waitFor(t, 5*time.Second, "the stale agent row to be deleted", func() bool {
		stored, err := logic.GetAllAgents()
		if err != nil {
			return false
		}
		return len(stored) == 1 && stored[0].Id == live.id
	})

	resp, err := client.StartTestExecution(ctx, &pb.StartTestExecutionRequest{
		UseCaseId:              "uc2",
		TargetId:               "test-target",
		SimulatedUsersPerAgent: 3,
		AgentSelection:         "all",
	})
	if err != nil {
		t.Fatalf("StartTestExecution failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("StartTestExecution was rejected: %s", resp.Message)
	}

	waitFor(t, 5*time.Second, "the live agent to receive the start command", func() bool {
		return live.receivedStartTests() == 1
	})

	if got := dying.receivedStartTests(); got != 0 {
		t.Errorf("expected the reaped agent to receive no start command, got %d", got)
	}
}
