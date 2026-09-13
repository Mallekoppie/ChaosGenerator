package agent

import (
	"log"

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
	case *pb.AgentCommand_StartTest:
		return h.handleStartTest(cmd.CommandId, command.StartTest)
	case *pb.AgentCommand_StopTest:
		return h.handleStopTest(cmd.CommandId, command.StopTest)
	case *pb.AgentCommand_GetVersion:
		return h.handleGetVersion(cmd.CommandId, command.GetVersion)
	case *pb.AgentCommand_HealthCheck:
		return h.handleHealthCheck(cmd.CommandId, command.HealthCheck)
	default:
		response.Error = "Unknown command type"
		return response
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
		cmd.Sni,
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

func (h *CommandHandler) handleHealthCheck(commandId string, cmd *pb.HealthCheckCommand) *pb.CommandResponse {
	return &pb.CommandResponse{
		CommandId: commandId,
		Success:   true,
	}
}
