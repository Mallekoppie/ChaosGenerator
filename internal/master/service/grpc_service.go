package service

import (
	"context"
	"errors"
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

// TestGroup operations
func (s *ChaosMasterServer) GetAllTestGroups(ctx context.Context, req *contracts.GetAllTestGroupsRequest) (*contracts.GetAllTestGroupsResponse, error) {
	platform.Log.Info("Received GetAllTestGroups request")

	testGroups, err := logic.GetAllTestGroups()
	if err != nil {
		platform.Log.Error("Error getting all test groups", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get test groups: %v", err)
	}

	pbTestGroups := make([]*contracts.TestGroup, len(testGroups))
	for i, group := range testGroups {
		pbTestGroups[i] = testGroupModelToProto(group)
	}

	return &contracts.GetAllTestGroupsResponse{TestGroups: pbTestGroups}, nil
}

func (s *ChaosMasterServer) GetTestGroup(ctx context.Context, req *contracts.GetTestGroupRequest) (*contracts.GetTestGroupResponse, error) {
	platform.Log.Info("Received GetTestGroup request", zap.String("id", req.Id))

	if len(req.Id) < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	group, err := logic.GetTestGroup(req.Id)
	if err != nil && err == platform.ErrNoEntryFoundInDB {
		return nil, status.Errorf(codes.NotFound, "test group not found")
	} else if err != nil {
		platform.Log.Error("Error getting test group", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get test group: %v", err)
	}

	return &contracts.GetTestGroupResponse{TestGroup: testGroupModelToProto(group)}, nil
}

func (s *ChaosMasterServer) AddTestGroup(ctx context.Context, req *contracts.AddTestGroupRequest) (*contracts.AddTestGroupResponse, error) {
	platform.Log.Info("Received AddTestGroup request", zap.String("name", req.TestGroup.Name))

	if req.TestGroup == nil {
		return nil, status.Errorf(codes.InvalidArgument, "test group is required")
	}

	group := testGroupProtoToModel(req.TestGroup)
	err := logic.CreateTestGroup(group)
	if err != nil {
		platform.Log.Error("Error creating test group", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create test group: %v", err)
	}

	return &contracts.AddTestGroupResponse{Success: true}, nil
}

func (s *ChaosMasterServer) UpdateTestGroup(ctx context.Context, req *contracts.UpdateTestGroupRequest) (*contracts.UpdateTestGroupResponse, error) {
	platform.Log.Info("Received UpdateTestGroup request", zap.String("id", req.TestGroup.Id))

	if req.TestGroup == nil {
		return nil, status.Errorf(codes.InvalidArgument, "test group is required")
	}

	group := testGroupProtoToModel(req.TestGroup)
	err := logic.UpdateTestGroup(group)
	if err != nil {
		platform.Log.Error("Error updating test group", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update test group: %v", err)
	}

	return &contracts.UpdateTestGroupResponse{Success: true}, nil
}

func (s *ChaosMasterServer) DeleteTestGroup(ctx context.Context, req *contracts.DeleteTestGroupRequest) (*contracts.DeleteTestGroupResponse, error) {
	platform.Log.Info("Received DeleteTestGroup request", zap.String("id", req.Id))

	if len(req.Id) < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	err := logic.DeleteTestGroup(req.Id)
	if err != nil {
		platform.Log.Error("Error deleting test group", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete test group: %v", err)
	}

	return &contracts.DeleteTestGroupResponse{Success: true}, nil
}

// TestCollection operations
func (s *ChaosMasterServer) AddTestCollection(ctx context.Context, req *contracts.AddTestCollectionRequest) (*contracts.AddTestCollectionResponse, error) {
	platform.Log.Info("Received AddTestCollection request", zap.String("name", req.TestCollection.Name))

	if req.TestCollection == nil {
		return nil, status.Errorf(codes.InvalidArgument, "test collection is required")
	}

	collection := testCollectionProtoToModel(req.TestCollection)
	err := logic.AddTestCollection(collection)
	if err != nil && err == logic.ErrParentTestGroupDoesNotExist {
		return &contracts.AddTestCollectionResponse{Success: false, Error: err.Error()}, nil
	} else if err != nil {
		platform.Log.Error("Error adding test collection", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to add test collection: %v", err)
	}

	return &contracts.AddTestCollectionResponse{Success: true}, nil
}

func (s *ChaosMasterServer) UpdateTestCollection(ctx context.Context, req *contracts.UpdateTestCollectionRequest) (*contracts.UpdateTestCollectionResponse, error) {
	platform.Log.Info("Received UpdateTestCollection request", zap.String("id", req.TestCollection.Id))

	if req.TestCollection == nil {
		return nil, status.Errorf(codes.InvalidArgument, "test collection is required")
	}

	collection := testCollectionProtoToModel(req.TestCollection)
	err := logic.UpdateTestCollection(collection)
	if err != nil && err == logic.ErrParentTestGroupDoesNotExist {
		return &contracts.UpdateTestCollectionResponse{Success: false, Error: err.Error()}, nil
	} else if err != nil {
		platform.Log.Error("Error updating test collection", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update test collection: %v", err)
	}

	return &contracts.UpdateTestCollectionResponse{Success: true}, nil
}

func (s *ChaosMasterServer) DeleteTestCollection(ctx context.Context, req *contracts.DeleteTestCollectionRequest) (*contracts.DeleteTestCollectionResponse, error) {
	platform.Log.Info("Received DeleteTestCollection request", zap.String("id", req.Id))

	if len(req.Id) < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	err := logic.DeleteTestCollection(req.Id)
	if err != nil {
		platform.Log.Error("Error deleting test collection", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete test collection: %v", err)
	}

	return &contracts.DeleteTestCollectionResponse{Success: true}, nil
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

func testGroupModelToProto(group models.TestGroup) *contracts.TestGroup {
	pbCollections := make([]*contracts.MasterTestCollection, len(group.TestCollections))
	for i, collection := range group.TestCollections {
		pbCollections[i] = testCollectionModelToProto(collection)
	}

	return &contracts.TestGroup{
		Id:              group.ID,
		Name:            group.Name,
		Description:     group.Description,
		TestCollections: pbCollections,
	}
}

func testGroupProtoToModel(pbGroup *contracts.TestGroup) models.TestGroup {
	collections := make([]models.TestCollection, len(pbGroup.TestCollections))
	for i, pbCollection := range pbGroup.TestCollections {
		collections[i] = testCollectionProtoToModel(pbCollection)
	}

	return models.TestGroup{
		ID:              pbGroup.Id,
		Name:            pbGroup.Name,
		Description:     pbGroup.Description,
		TestCollections: collections,
	}
}

func testCollectionModelToProto(collection models.TestCollection) *contracts.MasterTestCollection {
	pbTests := make([]*contracts.MasterTest, len(collection.Tests))
	for i, test := range collection.Tests {
		pbTests[i] = testModelToProto(test)
	}

	return &contracts.MasterTestCollection{
		Id:          collection.ID,
		Name:        collection.Name,
		Description: collection.Description,
		Tests:       pbTests,
		GroupId:     collection.GroupId,
	}
}

func testCollectionProtoToModel(pbCollection *contracts.MasterTestCollection) models.TestCollection {
	tests := make([]models.Test, len(pbCollection.Tests))
	for i, pbTest := range pbCollection.Tests {
		tests[i] = testProtoToModel(pbTest)
	}

	return models.TestCollection{
		ID:          pbCollection.Id,
		Name:        pbCollection.Name,
		Description: pbCollection.Description,
		Tests:       tests,
		GroupId:     pbCollection.GroupId,
	}
}

func testModelToProto(test models.Test) *contracts.MasterTest {
	pbHeaders := make([]*contracts.MasterHeader, len(test.Headers))
	for i, header := range test.Headers {
		pbHeaders[i] = &contracts.MasterHeader{
			Id:    header.ID,
			Name:  header.Name,
			Value: header.Value,
		}
	}

	return &contracts.MasterTest{
		Id:           test.ID,
		Name:         test.Name,
		Method:       test.Method,
		Url:          test.Url,
		Body:         test.Body,
		Headers:      pbHeaders,
		ResponseCode: test.ResponseCode,
		ResponseBody: test.ResponseBody,
	}
}

func testProtoToModel(pbTest *contracts.MasterTest) models.Test {
	headers := make([]models.Header, len(pbTest.Headers))
	for i, pbHeader := range pbTest.Headers {
		headers[i] = models.Header{
			ID:    pbHeader.Id,
			Name:  pbHeader.Name,
			Value: pbHeader.Value,
		}
	}

	return models.Test{
		ID:           pbTest.Id,
		Name:         pbTest.Name,
		Method:       pbTest.Method,
		Url:          pbTest.Url,
		Body:         pbTest.Body,
		Headers:      headers,
		ResponseCode: pbTest.ResponseCode,
		ResponseBody: pbTest.ResponseBody,
	}
}
