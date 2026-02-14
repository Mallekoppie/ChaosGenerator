package logic

import (
	"mallekoppie/ChaosGenerator/internal/master/models"
	"mallekoppie/ChaosGenerator/internal/master/repositories"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"

	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	consulAgentServiceName string = "ChaosAgent"
	consulAgentMetricsName string = "ChaosAgentMetrics"
)

var (
	ErrParentTestGroupDoesNotExist = errors.New("Parent TestGroup does not exist")
)

func GetAllAgents() (agents []models.Agent, err error) {
	platform.Log.Info("Getting all agents")

	agents, err = repositories.GetAllAgents()
	if err != nil {
		platform.Log.Error("Unable to get agents: ", zap.Error(err))
		return agents, err
	}

	platform.Log.Info("Returning agents successfully")
	return agents, nil
}

func UpdateAgent(agent models.Agent) error {
	platform.Log.Info("Adding agent")
	// Register Agent
	err := repositories.UpdateAgent(agent)
	if err != nil {
		platform.Log.Error("Unable to register normal agent in consul: ", zap.Error(err))
		return err
	}
	// Register Metrics
	//err = repositories.UpdateChaosAgent(agent, consulAgentMetricsName, agent.MetricsPort)
	//if err != nil {
	//	platform.Log.Error("Unable to register metrics agent in consul: ", zap.Error(err))
	//	return err
	//}

	platform.Log.Info("Added agent successfully")
	return nil
}

func DeleteAgent(agent models.Agent) error {
	platform.Log.Info("Deleting agent")

	err := repositories.DeleteAgent(agent.Id)
	if err != nil {
		platform.Log.Error("Unable to delete normal agent in consul: ", zap.Error(err))
		return err
	}
	//
	//err = repositories.DeleteChaosAgent(agent, consulAgentMetricsName)
	//if err != nil {
	//	platform.Log.Error("Unable to delete metric agent in consul: ", zap.Error(err))
	//	return err
	//}

	platform.Log.Debug("Agent deleted")
	return nil

}

func RegisterAgent(hostname string, port int, metricsPort int, version string) (string, error) {
	platform.Log.Info("Registering new agent",
		zap.String("hostname", hostname),
		zap.Int("port", port),
		zap.String("version", version))

	// Generate unique agent ID
	agentId := uuid.New().String()

	// Create agent model with initial status
	agent := models.Agent{
		Id:          agentId,
		Host:        hostname,
		Port:        port,
		MetricsPort: metricsPort,
		Enabled:     true,
		Status:      fmt.Sprintf("Connected at %s (version: %s)", time.Now().Format(time.RFC3339), version),
	}

	// Save agent to repository
	err := repositories.UpdateAgent(agent)
	if err != nil {
		platform.Log.Error("Unable to register agent in database", zap.Error(err))
		return "", err
	}

	platform.Log.Info("Agent registered successfully", zap.String("agentId", agentId))
	return agentId, nil
}

func GetAllTestGroups() (tests []models.TestGroup, err error) {
	tests, err = repositories.GetAllTestGroups()
	if err != nil {
		return tests, err
	}

	return tests, nil
}

func GetTestGroup(id string) (group models.TestGroup, err error) {
	group, err = repositories.GetTestGroup(id)
	if err != nil {
		platform.Log.Error("Unable to get Test Group: ", zap.Error(err))
		return group, err
	}

	return group, err
}

func CreateTestGroup(group models.TestGroup) error {
	err := repositories.AddTestGroup(group)
	if err != nil {
		platform.Log.Error("Unable to create new TestGroup", zap.Error(err))
		return err
	}

	return nil
}

func UpdateTestGroup(group models.TestGroup) error {
	err := repositories.UpdateTestGroup(group)
	return err
}

func DeleteTestGroup(id string) error {
	err := repositories.DeleteTestGroup(id)
	return err
}

func AddTestCollection(test models.TestCollection) error {

	result, err := repositories.DoesTestGroupExist(test.GroupId)
	if err != nil {
		platform.Log.Error("Error verifying if the test group exists when adding test collection", zap.Error(err))
		return err
	}

	if result == false {
		platform.Log.Warn("Parent Test Group must exist when adding a test collection")
		return ErrParentTestGroupDoesNotExist
	}

	err = repositories.AddTestCollection(test)

	return err
}

func UpdateTestCollection(test models.TestCollection) error {
	result, err := repositories.DoesTestGroupExist(test.GroupId)
	if err != nil {
		platform.Log.Error("Error verifying if the test group exists when adding test collection", zap.Error(err))
		return err
	}

	if result == false {
		platform.Log.Warn("Parent Test Group must exist when adding a test collection")
		return ErrParentTestGroupDoesNotExist
	}

	err = repositories.UpdateTestCollection(test)

	return err
}

func DeleteTestCollection(id string) error {
	err := repositories.DeleteTestCollection(id)
	return err
}

// Target management

func RegisterTarget(target models.Target) (string, error) {
	platform.Log.Info("Registering new target",
		zap.String("name", target.Name),
		zap.String("address", target.Address),
		zap.String("serviceType", target.ServiceType))

	// Generate unique target ID if not provided
	if target.ID == "" {
		target.ID = uuid.New().String()
	}

	// Save target to repository
	err := repositories.AddTarget(target)
	if err != nil {
		platform.Log.Error("Unable to register target in database", zap.Error(err))
		return "", err
	}

	platform.Log.Info("Target registered successfully", zap.String("targetId", target.ID))
	return target.ID, nil
}

func UpdateTarget(target models.Target) error {
	platform.Log.Info("Updating target", zap.String("id", target.ID))

	err := repositories.UpdateTarget(target)
	if err != nil {
		platform.Log.Error("Unable to update target", zap.Error(err))
		return err
	}

	platform.Log.Info("Target updated successfully")
	return nil
}

func DeleteTarget(id string) error {
	platform.Log.Info("Deleting target", zap.String("id", id))

	err := repositories.DeleteTarget(id)
	if err != nil {
		platform.Log.Error("Unable to delete target", zap.Error(err))
		return err
	}

	platform.Log.Debug("Target deleted")
	return nil
}

func GetAllTargets() ([]models.Target, error) {
	platform.Log.Info("Getting all targets")

	targets, err := repositories.GetAllTargets()
	if err != nil {
		platform.Log.Error("Unable to get targets", zap.Error(err))
		return nil, err
	}

	platform.Log.Info("Returning targets successfully")
	return targets, nil
}

func GetTarget(id string) (models.Target, error) {
	platform.Log.Info("Getting target", zap.String("id", id))

	target, err := repositories.GetTarget(id)
	if err != nil {
		platform.Log.Error("Unable to get target", zap.Error(err))
		return models.Target{}, err
	}

	return target, nil
}
