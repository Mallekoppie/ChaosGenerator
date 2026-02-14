package manager

import (
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/tkanos/gonfig"

	pb "mallekoppie/ChaosGenerator/internal/contracts"
	"mallekoppie/ChaosGenerator/internal/master/service"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
)

var (
	ErrorNoAgentWithThatId = errors.New("No agent with that Id")
	ErrorAgentNotConnected = errors.New("Agent not connected")
)

// GetAllConnectedAgents returns all agents that are currently connected
func GetAllConnectedAgents() []string {
	cm := service.GetConnectionManager()
	conns := cm.GetAllConnections()

	agentIds := make([]string, len(conns))
	for i, conn := range conns {
		agentIds[i] = conn.AgentId
	}

	return agentIds
}

// IsAgentConnected checks if an agent is currently connected
func IsAgentConnected(agentId string) bool {
	cm := service.GetConnectionManager()
	_, ok := cm.GetConnection(agentId)
	return ok
}

func GetTest(testName string) (*pb.TestCollection, error) {
	configuration := &pb.TestCollection{}
	err := gonfig.GetConf("./tests/"+testName+".json", configuration)

	if err != nil {
		log.Printf("Error reading config: %v", err)
		return nil, err
	}

	return configuration, nil
}

// SendAddTestsCommand sends add tests command to an agent
func SendAddTestsCommand(agentId string, testCollection *pb.TestCollection) error {
	if !IsAgentConnected(agentId) {
		return ErrorAgentNotConnected
	}

	cm := service.GetConnectionManager()

	// Convert TestCollection to AddTestsCommand format
	tests := make([]*pb.AgentTest, len(testCollection.Tests))
	for i, test := range testCollection.Tests {
		headers := make([]*pb.AgentHeader, len(test.Headers))
		for j, header := range test.Headers {
			headers[j] = &pb.AgentHeader{
				Name:  header.Name,
				Value: header.Value,
			}
		}
		tests[i] = &pb.AgentTest{
			Name:         test.Name,
			Method:       test.Method,
			Url:          test.Url,
			Body:         test.Body,
			Headers:      headers,
			ResponseCode: test.ResponseCode,
			ResponseBody: test.ResponseBody,
		}
	}

	command := &pb.AgentCommand{
		CommandId: uuid.New().String(),
		Command: &pb.AgentCommand_AddTests{
			AddTests: &pb.AddTestsCommand{
				Name:  testCollection.Name,
				Tests: tests,
			},
		},
	}

	_, err := cm.SendCommand(agentId, command)
	if err != nil {
		platform.Log.Error("Error sending AddTests command", zap.String("agentId", agentId), zap.Error(err))
		return err
	}

	platform.Log.Info("AddTests command sent", zap.String("agentId", agentId))
	return nil
}

// SendStartTestCommand sends start test command to an agent
func SendStartTestCommand(agentId string, testCollectionName string, simulatedUsers int32) error {
	if !IsAgentConnected(agentId) {
		return ErrorAgentNotConnected
	}

	cm := service.GetConnectionManager()

	command := &pb.AgentCommand{
		CommandId: uuid.New().String(),
		Command: &pb.AgentCommand_StartTest{
			StartTest: &pb.StartTestCommand{
				TestExecutionId: testCollectionName,
				SimulatedUsers:  simulatedUsers,
				// TODO: Add use case and target info from test execution
			},
		},
	}

	_, err := cm.SendCommand(agentId, command)
	if err != nil {
		platform.Log.Error("Error sending StartTest command", zap.String("agentId", agentId), zap.Error(err))
		return err
	}

	platform.Log.Info("StartTest command sent", zap.String("agentId", agentId))
	return nil
}

// SendStopTestCommand sends stop test command to an agent
func SendStopTestCommand(agentId string, testName string) error {
	if !IsAgentConnected(agentId) {
		return ErrorAgentNotConnected
	}

	cm := service.GetConnectionManager()

	command := &pb.AgentCommand{
		CommandId: uuid.New().String(),
		Command: &pb.AgentCommand_StopTest{
			StopTest: &pb.StopTestCommand{
				TestExecutionId: testName,
			},
		},
	}

	_, err := cm.SendCommand(agentId, command)
	if err != nil {
		platform.Log.Error("Error sending StopTest command", zap.String("agentId", agentId), zap.Error(err))
		return err
	}

	platform.Log.Info("StopTest command sent", zap.String("agentId", agentId))
	return nil
}

// SendUpdateTestCommand sends update test command to an agent
func SendUpdateTestCommand(agentId string, testCollectionName string, simulatedUsers int32) error {
	if !IsAgentConnected(agentId) {
		return ErrorAgentNotConnected
	}

	cm := service.GetConnectionManager()

	command := &pb.AgentCommand{
		CommandId: uuid.New().String(),
		Command: &pb.AgentCommand_UpdateTest{
			UpdateTest: &pb.UpdateTestCommand{
				TestCollectionName: testCollectionName,
				SimulatedUsers:     simulatedUsers,
			},
		},
	}

	_, err := cm.SendCommand(agentId, command)
	if err != nil {
		platform.Log.Error("Error sending UpdateTest command", zap.String("agentId", agentId), zap.Error(err))
		return err
	}

	platform.Log.Info("UpdateTest command sent", zap.String("agentId", agentId))
	return nil
}

// SendGetStatusCommand sends get status command to an agent
func SendGetStatusCommand(agentId string) error {
	if !IsAgentConnected(agentId) {
		return ErrorAgentNotConnected
	}

	cm := service.GetConnectionManager()

	command := &pb.AgentCommand{
		CommandId: uuid.New().String(),
		Command: &pb.AgentCommand_GetStatus{
			GetStatus: &pb.GetStatusCommand{},
		},
	}

	_, err := cm.SendCommand(agentId, command)
	if err != nil {
		platform.Log.Error("Error sending GetStatus command", zap.String("agentId", agentId), zap.Error(err))
		return err
	}

	platform.Log.Info("GetStatus command sent", zap.String("agentId", agentId))
	return nil
}

// SendGetVersionCommand sends get version command to an agent
func SendGetVersionCommand(agentId string) error {
	if !IsAgentConnected(agentId) {
		return ErrorAgentNotConnected
	}

	cm := service.GetConnectionManager()

	command := &pb.AgentCommand{
		CommandId: uuid.New().String(),
		Command: &pb.AgentCommand_GetVersion{
			GetVersion: &pb.GetVersionCommand{},
		},
	}

	_, err := cm.SendCommand(agentId, command)
	if err != nil {
		platform.Log.Error("Error sending GetVersion command", zap.String("agentId", agentId), zap.Error(err))
		return err
	}

	platform.Log.Info("GetVersion command sent", zap.String("agentId", agentId))
	return nil
}

// SendDeleteTestsCommand sends delete tests command to an agent
func SendDeleteTestsCommand(agentId string) error {
	if !IsAgentConnected(agentId) {
		return ErrorAgentNotConnected
	}

	cm := service.GetConnectionManager()

	command := &pb.AgentCommand{
		CommandId: uuid.New().String(),
		Command: &pb.AgentCommand_DeleteTests{
			DeleteTests: &pb.DeleteTestsCommand{},
		},
	}

	_, err := cm.SendCommand(agentId, command)
	if err != nil {
		platform.Log.Error("Error sending DeleteTests command", zap.String("agentId", agentId), zap.Error(err))
		return err
	}

	platform.Log.Info("DeleteTests command sent", zap.String("agentId", agentId))
	return nil
}

// SendHealthCheckCommand sends health check command to an agent
func SendHealthCheckCommand(agentId string) error {
	if !IsAgentConnected(agentId) {
		return ErrorAgentNotConnected
	}

	cm := service.GetConnectionManager()

	command := &pb.AgentCommand{
		CommandId: uuid.New().String(),
		Command: &pb.AgentCommand_HealthCheck{
			HealthCheck: &pb.HealthCheckCommand{},
		},
	}

	_, err := cm.SendCommand(agentId, command)
	if err != nil {
		platform.Log.Error("Error sending HealthCheck command", zap.String("agentId", agentId), zap.Error(err))
		return err
	}

	platform.Log.Info("HealthCheck command sent", zap.String("agentId", agentId))
	return nil
}
