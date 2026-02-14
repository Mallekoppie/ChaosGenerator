package logic

import (
	"mallekoppie/ChaosGenerator/internal/master/models"
	"mallekoppie/ChaosGenerator/internal/master/repositories"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"

	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	consulAgentServiceName string = "ChaosAgent"
	consulAgentMetricsName string = "ChaosAgentMetrics"
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

func ClearAllAgents() (int, error) {
	platform.Log.Info("Clearing all agents from database")

	// Get count before deleting
	agents, err := repositories.GetAllAgents()
	if err != nil {
		platform.Log.Error("Unable to get agent count: ", zap.Error(err))
		return 0, err
	}
	count := len(agents)

	err = repositories.DeleteAllAgents()
	if err != nil {
		platform.Log.Error("Unable to delete all agents: ", zap.Error(err))
		return 0, err
	}

	platform.Log.Info("All agents cleared", zap.Int("count", count))
	return count, nil
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
