package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mallekoppie/ChaosGenerator/internal/contracts"
	"mallekoppie/ChaosGenerator/internal/master/logic"
	"mallekoppie/ChaosGenerator/internal/master/models"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAgentNotConnected = errors.New("agent not connected")
	ErrInvalidAgentId    = errors.New("invalid agent ID")
)

type ChaosMasterServer struct {
	contracts.UnimplementedChaosMasterServer
}

func (s *ChaosMasterServer) Register(server *grpc.Server) {
	contracts.RegisterChaosMasterServer(server, s)
}

// Agent operations
func (s *ChaosMasterServer) GetAllAgents(ctx context.Context, req *contracts.GetAllAgentsRequest) (*contracts.GetAllAgentsResponse, error) {
	platform.Log.Info("Received GetAllAgents request")

	agents, err := logic.GetAllAgents()
	if err != nil {
		platform.Log.Error("Error getting all agents", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get agents: %v", err)
	}

	pbAgents := make([]*contracts.Agent, len(agents))
	for i, agent := range agents {
		pbAgents[i] = agentModelToProto(agent)
	}

	return &contracts.GetAllAgentsResponse{Agents: pbAgents}, nil
}

func (s *ChaosMasterServer) UpdateAgent(ctx context.Context, req *contracts.UpdateAgentRequest) (*contracts.UpdateAgentResponse, error) {
	platform.Log.Info("Received UpdateAgent request", zap.String("agentId", req.Agent.Id))

	if req.Agent == nil {
		return nil, status.Errorf(codes.InvalidArgument, "agent is required")
	}

	agent := agentProtoToModel(req.Agent)
	err := logic.UpdateAgent(agent)
	if err != nil {
		platform.Log.Error("Error updating agent", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update agent: %v", err)
	}

	return &contracts.UpdateAgentResponse{Success: true}, nil
}

func (s *ChaosMasterServer) DeleteAgent(ctx context.Context, req *contracts.DeleteAgentRequest) (*contracts.DeleteAgentResponse, error) {
	platform.Log.Info("Received DeleteAgent request", zap.String("agentId", req.Agent.Id))

	if req.Agent == nil {
		return nil, status.Errorf(codes.InvalidArgument, "agent is required")
	}

	agent := agentProtoToModel(req.Agent)
	err := logic.DeleteAgent(agent)
	if err != nil {
		platform.Log.Error("Error deleting agent", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete agent: %v", err)
	}

	return &contracts.DeleteAgentResponse{Success: true}, nil
}

func (s *ChaosMasterServer) ClearAllAgents(ctx context.Context, req *contracts.ClearAllAgentsRequest) (*contracts.ClearAllAgentsResponse, error) {
	platform.Log.Info("Received ClearAllAgents request")

	// First, close all active connections
	connMgr := GetConnectionManager()
	activeAgentIds := connMgr.GetAllAgentIds()

	// Note: Closing connections will trigger RemoveConnection which deletes from DB
	// So we clear the database after to ensure everything is cleaned up

	// Clear from database
	count, err := logic.ClearAllAgents()
	if err != nil {
		platform.Log.Error("Error clearing all agents", zap.Error(err))
		return &contracts.ClearAllAgentsResponse{
			Success:      false,
			Message:      fmt.Sprintf("Failed to clear agents: %v", err),
			DeletedCount: 0,
		}, nil
	}

	platform.Log.Info("All agents cleared from database",
		zap.Int("deletedCount", count),
		zap.Int("activeConnections", len(activeAgentIds)))

	return &contracts.ClearAllAgentsResponse{
		Success:      true,
		Message:      fmt.Sprintf("Successfully cleared %d agent(s) from database", count),
		DeletedCount: int32(count),
	}, nil
}

func (s *ChaosMasterServer) RegisterAgent(ctx context.Context, req *contracts.RegisterAgentRequest) (*contracts.RegisterAgentResponse, error) {
	platform.Log.Info("Received RegisterAgent request",
		zap.String("hostname", req.Hostname),
		zap.Int32("port", req.Port),
		zap.String("version", req.Version))

	if req.Hostname == "" {
		return nil, status.Errorf(codes.InvalidArgument, "hostname is required")
	}
	if req.Port <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "valid port is required")
	}

	agentId, err := logic.RegisterAgent(req.Hostname, int(req.Port), int(req.MetricsPort), req.Version)
	if err != nil {
		platform.Log.Error("Error registering agent", zap.Error(err))
		return &contracts.RegisterAgentResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	platform.Log.Info("Agent registered successfully", zap.String("agentId", agentId))
	return &contracts.RegisterAgentResponse{
		AgentId: agentId,
		Success: true,
		Message: "Agent registered successfully",
	}, nil
}

// Target operations

func (s *ChaosMasterServer) RegisterTarget(ctx context.Context, req *contracts.RegisterTargetRequest) (*contracts.RegisterTargetResponse, error) {
	platform.Log.Info("Received RegisterTarget request", zap.String("name", req.Target.Name))

	if req.Target == nil {
		return nil, status.Errorf(codes.InvalidArgument, "target is required")
	}
	if req.Target.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "target name is required")
	}
	if req.Target.Address == "" {
		return nil, status.Errorf(codes.InvalidArgument, "target address is required")
	}

	target := targetProtoToModel(req.Target)
	targetId, err := logic.RegisterTarget(target)
	if err != nil {
		platform.Log.Error("Error registering target", zap.Error(err))
		return &contracts.RegisterTargetResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	platform.Log.Info("Target registered successfully", zap.String("targetId", targetId))
	return &contracts.RegisterTargetResponse{
		TargetId: targetId,
		Success:  true,
		Message:  "Target registered successfully",
	}, nil
}

func (s *ChaosMasterServer) UpdateTarget(ctx context.Context, req *contracts.UpdateTargetRequest) (*contracts.UpdateTargetResponse, error) {
	platform.Log.Info("Received UpdateTarget request", zap.String("id", req.Target.Id))

	if req.Target == nil {
		return nil, status.Errorf(codes.InvalidArgument, "target is required")
	}
	if req.Target.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "target ID is required")
	}

	target := targetProtoToModel(req.Target)
	err := logic.UpdateTarget(target)
	if err != nil {
		platform.Log.Error("Error updating target", zap.Error(err))
		return &contracts.UpdateTargetResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &contracts.UpdateTargetResponse{
		Success: true,
		Message: "Target updated successfully",
	}, nil
}

func (s *ChaosMasterServer) DeleteTarget(ctx context.Context, req *contracts.DeleteTargetRequest) (*contracts.DeleteTargetResponse, error) {
	platform.Log.Info("Received DeleteTarget request", zap.String("id", req.Id))

	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "target ID is required")
	}

	err := logic.DeleteTarget(req.Id)
	if err != nil {
		platform.Log.Error("Error deleting target", zap.Error(err))
		return &contracts.DeleteTargetResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &contracts.DeleteTargetResponse{
		Success: true,
		Message: "Target deleted successfully",
	}, nil
}

func (s *ChaosMasterServer) GetAllTargets(ctx context.Context, req *contracts.GetAllTargetsRequest) (*contracts.GetAllTargetsResponse, error) {
	platform.Log.Info("Received GetAllTargets request")

	targets, err := logic.GetAllTargets()
	if err != nil {
		platform.Log.Error("Error getting all targets", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get targets: %v", err)
	}

	pbTargets := make([]*contracts.Target, len(targets))
	for i, target := range targets {
		pbTargets[i] = targetModelToProto(target)
	}

	return &contracts.GetAllTargetsResponse{Targets: pbTargets}, nil
}

// Use case operations

func (s *ChaosMasterServer) GetUseCases(ctx context.Context, req *contracts.GetUseCasesRequest) (*contracts.GetUseCasesResponse, error) {
	platform.Log.Info("Received GetUseCases request")

	// Hardcoded use cases
	useCases := getAllUseCases()

	return &contracts.GetUseCasesResponse{UseCases: useCases}, nil
}

// Helper function to get all use cases
func getAllUseCases() []*contracts.UseCase {
	return []*contracts.UseCase{
		{Id: "uc1", Name: "Health Check", Path: "/health", Method: "GET", Description: "Minimal overhead health check endpoint"},
		{Id: "uc2", Name: "Status Endpoint", Path: "/api/v1/status", Method: "GET", Description: "Status endpoint with lightweight JSON metadata"},
		{Id: "uc3", Name: "Simple Query", Path: "/api/v1/users/", Method: "GET", Description: "Simple query operation with small payload (1-2KB)"},
		{Id: "uc4", Name: "Create Resource", Path: "/api/v1/users", Method: "POST", Description: "Create resource operation with request/response (1-2KB)"},
		{Id: "uc5", Name: "List Collection", Path: "/api/v1/users", Method: "GET", Description: "List/collection query with pagination (50-100KB)"},
		{Id: "uc6", Name: "Bulk Update", Path: "/api/v1/users/bulk", Method: "PATCH", Description: "Bulk update operation with larger payloads (100-200KB)"},
		{Id: "uc7", Name: "Report Generation", Path: "/api/v1/reports/generate", Method: "POST", Description: "CPU-intensive report generation (500KB-1MB response)"},
		{Id: "uc8", Name: "File Upload", Path: "/api/v1/documents/upload", Method: "POST", Description: "Large request payload simulation (10-25MB)"},
		{Id: "uc9", Name: "Data Export", Path: "/api/v1/transactions/export", Method: "GET", Description: "Large response payload and streaming (25-50MB)"},
		{Id: "uc10", Name: "Network Saturation", Path: "/api/v1/data/process", Method: "POST", Description: "Network bandwidth saturation testing (100-250MB)"},
		{Id: "uc11", Name: "Memory Pressure", Path: "/api/v1/analytics/stream", Method: "POST", Description: "Long-running operations for memory leak detection (10-20KB)"},
	}
}

// Helper function to get use case name by ID
func getUseCaseName(useCaseId string) string {
	useCases := getAllUseCases()
	for _, uc := range useCases {
		if uc.Id == useCaseId {
			return uc.Name
		}
	}
	return "Unknown Use Case"
}

// Test execution operations

func (s *ChaosMasterServer) StartTestExecution(ctx context.Context, req *contracts.StartTestExecutionRequest) (*contracts.StartTestExecutionResponse, error) {
	platform.Log.Info("Received StartTestExecution request",
		zap.String("useCaseId", req.UseCaseId),
		zap.String("targetId", req.TargetId),
		zap.Int32("users", req.SimulatedUsersPerAgent))

	// Validate request
	if req.UseCaseId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "use case ID is required")
	}
	if req.TargetId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "target ID is required")
	}
	if req.SimulatedUsersPerAgent <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "simulated users per agent must be greater than 0")
	}

	// Get target details
	target, err := logic.GetTarget(req.TargetId)
	if err != nil {
		platform.Log.Error("Error getting target", zap.Error(err))
		return &contracts.StartTestExecutionResponse{
			Success: false,
			Message: fmt.Sprintf("Target not found: %v", err),
		}, nil
	}

	// Generate test execution ID
	testExecutionId := fmt.Sprintf("test-%d", time.Now().Unix())

	// Get connected agents
	connMgr := GetConnectionManager()
	var agentIds []string

	if req.AgentSelection == "all" || req.AgentSelection == "" {
		agentIds = connMgr.GetAllAgentIds()
	} else {
		// Parse comma-separated agent IDs
		agentIds = parseAgentSelection(req.AgentSelection)
	}

	if len(agentIds) == 0 {
		return &contracts.StartTestExecutionResponse{
			Success: false,
			Message: "No agents available or selected",
		}, nil
	}

	// Send StartTestCommand to all selected agents
	successCount := 0
	for _, agentId := range agentIds {
		command := &contracts.AgentCommand{
			CommandId: fmt.Sprintf("cmd-%s-%s", testExecutionId, agentId),
			Command: &contracts.AgentCommand_StartTest{
				StartTest: &contracts.StartTestCommand{
					TestExecutionId:        testExecutionId,
					UseCaseId:              req.UseCaseId,
					TargetAddress:          target.Address,
					TargetProtocol:         target.Protocol,
					ConnectionReuseEnabled: target.ConnectionReuseEnabled,
					SimulatedUsers:         req.SimulatedUsersPerAgent,
				},
			},
		}

		_, err := connMgr.SendCommand(agentId, command)
		if err != nil {
			platform.Log.Error("Error sending start command to agent",
				zap.String("agentId", agentId),
				zap.Error(err))
		} else {
			successCount++
		}
	}

	if successCount == 0 {
		return &contracts.StartTestExecutionResponse{
			Success: false,
			Message: "Failed to start test on any agent",
		}, nil
	}

	platform.Log.Info("Test execution started",
		zap.String("testExecutionId", testExecutionId),
		zap.Int("agents", successCount))

	// Track the running test
	tracker := GetTestExecutionTracker()
	runningTest := &RunningTest{
		TestExecutionId:        testExecutionId,
		UseCaseId:              req.UseCaseId,
		UseCaseName:            getUseCaseName(req.UseCaseId),
		TargetId:               target.ID,
		TargetName:             target.Name,
		SimulatedUsersPerAgent: req.SimulatedUsersPerAgent,
		AgentIds:               agentIds[:successCount],
		StartTime:              time.Now(),
	}
	tracker.AddTest(runningTest)

	return &contracts.StartTestExecutionResponse{
		TestExecutionId: testExecutionId,
		Success:         true,
		Message:         fmt.Sprintf("Test started on %d agent(s)", successCount),
	}, nil
}

func (s *ChaosMasterServer) StopTestExecution(ctx context.Context, req *contracts.StopTestExecutionRequest) (*contracts.StopTestExecutionResponse, error) {
	platform.Log.Info("Received StopTestExecution request", zap.String("testExecutionId", req.TestExecutionId))

	if req.TestExecutionId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "test execution ID is required")
	}

	// Get the running test details
	tracker := GetTestExecutionTracker()
	runningTest, exists := tracker.GetTest(req.TestExecutionId)
	if !exists {
		return &contracts.StopTestExecutionResponse{
			Success: false,
			Message: "Test execution not found or already stopped",
		}, nil
	}

	// Send StopTestCommand only to agents running this test
	connMgr := GetConnectionManager()
	agentIds := runningTest.AgentIds

	if len(agentIds) == 0 {
		return &contracts.StopTestExecutionResponse{
			Success: false,
			Message: "No agents running this test",
		}, nil
	}

	successCount := 0
	for _, agentId := range agentIds {
		command := &contracts.AgentCommand{
			CommandId: fmt.Sprintf("cmd-stop-%s-%s", req.TestExecutionId, agentId),
			Command: &contracts.AgentCommand_StopTest{
				StopTest: &contracts.StopTestCommand{
					TestExecutionId: req.TestExecutionId,
				},
			},
		}

		_, err := connMgr.SendCommand(agentId, command)
		if err != nil {
			platform.Log.Error("Error sending stop command to agent",
				zap.String("agentId", agentId),
				zap.Error(err))
		} else {
			successCount++
		}
	}

	platform.Log.Info("Stop command sent",
		zap.String("testExecutionId", req.TestExecutionId),
		zap.Int("agents", successCount))

	// Remove test from tracker
	tracker.RemoveTest(req.TestExecutionId)

	return &contracts.StopTestExecutionResponse{
		Success: true,
		Message: fmt.Sprintf("Stop command sent to %d agent(s)", successCount),
	}, nil
}

func (s *ChaosMasterServer) GetRunningTests(ctx context.Context, req *contracts.GetRunningTestsRequest) (*contracts.GetRunningTestsResponse, error) {
	platform.Log.Info("Received GetRunningTests request")

	tracker := GetTestExecutionTracker()
	runningTests := tracker.GetAllTests()

	// Convert to protobuf format
	pbTests := make([]*contracts.RunningTestExecution, len(runningTests))
	for i, test := range runningTests {
		pbTests[i] = test.ToProto()
	}

	platform.Log.Info("Returning running tests", zap.Int("count", len(pbTests)))

	return &contracts.GetRunningTestsResponse{
		Tests: pbTests,
	}, nil
}

// Helper functions

func parseAgentSelection(selection string) []string {
	if selection == "" {
		return []string{}
	}

	parts := strings.Split(selection, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func targetModelToProto(target models.Target) *contracts.Target {
	return &contracts.Target{
		Id:                     target.ID,
		Name:                   target.Name,
		Address:                target.Address,
		ServiceType:            target.ServiceType,
		Protocol:               target.Protocol,
		ConnectionReuseEnabled: target.ConnectionReuseEnabled,
	}
}

// Conversion functions
func agentModelToProto(agent models.Agent) *contracts.Agent {
	return &contracts.Agent{
		Id:          agent.Id,
		Host:        agent.Host,
		Port:        int32(agent.Port),
		MetricsPort: int32(agent.MetricsPort),
		Enabled:     agent.Enabled,
		Status:      agent.Status,
	}
}

func agentProtoToModel(pbAgent *contracts.Agent) models.Agent {
	return models.Agent{
		Id:          pbAgent.Id,
		Host:        pbAgent.Host,
		Port:        int(pbAgent.Port),
		MetricsPort: int(pbAgent.MetricsPort),
		Enabled:     pbAgent.Enabled,
		Status:      pbAgent.Status,
	}
}

// Target conversion helpers

func targetProtoToModel(pbTarget *contracts.Target) models.Target {
	return models.Target{
		ID:                     pbTarget.Id,
		Name:                   pbTarget.Name,
		Address:                pbTarget.Address,
		ServiceType:            pbTarget.ServiceType,
		Protocol:               pbTarget.Protocol,
		ConnectionReuseEnabled: pbTarget.ConnectionReuseEnabled,
	}
}
