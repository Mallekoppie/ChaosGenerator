package agent

import (
	"log"

	core "mallekoppie/ChaosGenerator/internal/agent/go"
	pb "mallekoppie/ChaosGenerator/internal/contracts"
)

// CommandHandler processes commands from the master server
type CommandHandler struct{}

// NewCommandHandler creates a new command handler
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{}
}

// HandleCommand processes a command from the master and returns a response
func (h *CommandHandler) HandleCommand(cmd *pb.AgentCommand) *pb.CommandResponse {
	response := &pb.CommandResponse{
		CommandId: cmd.CommandId,
		Success:   false,
	}

	switch command := cmd.Command.(type) {
	case *pb.AgentCommand_AddTests:
		return h.handleAddTests(cmd.CommandId, command.AddTests)
	case *pb.AgentCommand_StartTest:
		return h.handleStartTest(cmd.CommandId, command.StartTest)
	case *pb.AgentCommand_StopTest:
		return h.handleStopTest(cmd.CommandId, command.StopTest)
	case *pb.AgentCommand_UpdateTest:
		return h.handleUpdateTest(cmd.CommandId, command.UpdateTest)
	case *pb.AgentCommand_GetStatus:
		return h.handleGetStatus(cmd.CommandId, command.GetStatus)
	case *pb.AgentCommand_GetVersion:
		return h.handleGetVersion(cmd.CommandId, command.GetVersion)
	case *pb.AgentCommand_DeleteTests:
		return h.handleDeleteTests(cmd.CommandId, command.DeleteTests)
	case *pb.AgentCommand_HealthCheck:
		return h.handleHealthCheck(cmd.CommandId, command.HealthCheck)
	default:
		response.Error = "Unknown command type"
		return response
	}
}

func (h *CommandHandler) handleAddTests(commandId string, cmd *pb.AddTestsCommand) *pb.CommandResponse {
	log.Printf("Handling AddTests command: %s", cmd.Name)

	// Convert AgentTest to TestCollection format
	tests := make([]*pb.Test, len(cmd.Tests))
	for i, test := range cmd.Tests {
		headers := make([]*pb.Header, len(test.Headers))
		for j, header := range test.Headers {
			headers[j] = &pb.Header{
				Name:  header.Name,
				Value: header.Value,
			}
		}
		tests[i] = &pb.Test{
			Name:         test.Name,
			Method:       test.Method,
			Url:          test.Url,
			Body:         test.Body,
			Headers:      headers,
			ResponseCode: test.ResponseCode,
			ResponseBody: test.ResponseBody,
		}
	}

	testCollection := &pb.TestCollection{
		Name:  cmd.Name,
		Tests: tests,
	}

	core.WriteTestConfiguration(*testCollection)

	return &pb.CommandResponse{
		CommandId: commandId,
		Success:   true,
	}
}

func (h *CommandHandler) handleStartTest(commandId string, cmd *pb.StartTestCommand) *pb.CommandResponse {
	log.Printf("Handling StartTest command: %s with %d users, useCase: %s, target: %s",
		cmd.TestExecutionId, cmd.SimulatedUsers, cmd.UseCaseId, cmd.TargetAddress)

	// Use new test execution manager
	manager := GetTestExecutionManager()
	err := manager.StartTest(
		cmd.TestExecutionId,
		cmd.UseCaseId,
		cmd.TargetAddress,
		cmd.TargetProtocol,
		cmd.ConnectionReuseEnabled,
		int(cmd.SimulatedUsers),
	)

	if err != nil {
		log.Printf("Error starting test: %v", err)
		return &pb.CommandResponse{
			CommandId: commandId,
			Success:   false,
			Error:     err.Error(),
		}
	}

	return &pb.CommandResponse{
		CommandId: commandId,
		Success:   true,
	}
}

func (h *CommandHandler) handleStopTest(commandId string, cmd *pb.StopTestCommand) *pb.CommandResponse {
	log.Printf("Handling StopTest command: %s", cmd.TestExecutionId)

	// Use new test execution manager
	manager := GetTestExecutionManager()
	err := manager.StopTest(cmd.TestExecutionId)

	if err != nil {
		log.Printf("Error stopping test: %v", err)
		return &pb.CommandResponse{
			CommandId: commandId,
			Success:   false,
			Error:     err.Error(),
		}
	}

	return &pb.CommandResponse{
		CommandId: commandId,
		Success:   true,
	}
}

func (h *CommandHandler) handleUpdateTest(commandId string, cmd *pb.UpdateTestCommand) *pb.CommandResponse {
	log.Printf("Handling UpdateTest command: %d users", cmd.SimulatedUsers)

	err := core.CoreUpdateTest(cmd.SimulatedUsers)

	if err != nil {
		return &pb.CommandResponse{
			CommandId: commandId,
			Success:   false,
			Error:     err.Error(),
		}
	}

	return &pb.CommandResponse{
		CommandId: commandId,
		Success:   true,
	}
}

func (h *CommandHandler) handleGetStatus(commandId string, cmd *pb.GetStatusCommand) *pb.CommandResponse {
	log.Println("Handling GetStatus command")

	status := core.CoreGetTestStatus()

	return &pb.CommandResponse{
		CommandId: commandId,
		Success:   true,
		Result: &pb.CommandResponse_StatusResult{
			StatusResult: &pb.TestStatusResult{
				TestCollectionName:    status.TestCollectionName,
				RequestsExecuted:      status.RequestsExecuted,
				TransactionsPerSecond: status.TransactionsPerSecond,
				AverageExecutionTime:  status.AverageExecutionTime,
				Cpu:                   status.Cpu,
				SimulatedUsers:        status.SimulatedUsers,
				ErrorsPerSecond:       status.ErrorsPerSecond,
				ErrorsRaised:          status.ErrorsRaised,
				ExecutionTime:         status.ExecutionTime,
			},
		},
	}
}

func (h *CommandHandler) handleGetVersion(commandId string, cmd *pb.GetVersionCommand) *pb.CommandResponse {
	log.Println("Handling GetVersion command")

	return &pb.CommandResponse{
		CommandId: commandId,
		Success:   true,
		Result: &pb.CommandResponse_VersionResult{
			VersionResult: &pb.VersionResult{
				Version:  "2.0.0",
				Hostname: "Unknown",
			},
		},
	}
}

func (h *CommandHandler) handleDeleteTests(commandId string, cmd *pb.DeleteTestsCommand) *pb.CommandResponse {
	log.Println("Handling DeleteTests command")

	err := core.ClearTestsDirectory()

	if err != nil {
		return &pb.CommandResponse{
			CommandId: commandId,
			Success:   false,
			Error:     err.Error(),
		}
	}

	return &pb.CommandResponse{
		CommandId: commandId,
		Success:   true,
	}
}

func (h *CommandHandler) handleHealthCheck(commandId string, cmd *pb.HealthCheckCommand) *pb.CommandResponse {
	return &pb.CommandResponse{
		CommandId: commandId,
		Success:   true,
	}
}
