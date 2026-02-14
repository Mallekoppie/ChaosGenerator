package logic

import (
	"mallekoppie/ChaosGenerator/internal/master/models"
	"mallekoppie/ChaosGenerator/internal/master/repositories"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"

	"errors"
)

const (
	consulAgentServiceName string = "ChaosAgent"
	consulAgentMetricsName string = "ChaosAgentMetrics"
)

var (
	ErrParentTestGroupDoesNotExist = errors.New("Parent TestGroup does not exist")
)

func GetAllAgents() (agents []models.Agent, err error) {
	platform.Logger.Info("Getting all agents")

	agents, err = repositories.GetAllAgents()
	if err != nil {
		platform.Logger.Error("Unable to get agents: ", zap.Error(err))
		return agents, err
	}

	//agentmetrics, err := repositories.GetAllAgents(consulAgentMetricsName)
	//if err != nil {
	//	platform.Logger.Error("Unable to get agents: ", zap.Error(err))
	//	return agents, err
	//}
	//
	//// Add metrics port to agent
	//for index := range agents {
	//	for metricIndex := range agentmetrics {
	//		if agents[index].Id == agentmetrics[metricIndex].Id {
	//			agents[index].MetricsPort = agentmetrics[metricIndex].Port
	//		}
	//	}
	//}

	// Get Agent status
	//for index := range agents {
	//	agent := agents[index]
	//
	//	chaosAgent, err := manager.GetAgent(agent.Id)
	//	if err != nil {
	//		platform.Logger.Error("Unable to get agent for Id: ", zap.String("id", agent.Id))
	//		agents[index].Status = "error"
	//		continue
	//	}
	//
	//	alive := chaosAgent.IsAlive()
	//
	//	platform.Logger.Info("Agent IsAlive response", zap.Bool("alive",alive))
	//
	//	if alive == true {
	//		platform.Logger.Info("Setting online")
	//		agents[index].Status = "online"
	//	} else {
	//		platform.Logger.Info("Setting offline")
	//		agents[index].Status = "offline"
	//	}
	//}

	platform.Logger.Info("Returning agents successfully")
	return agents, nil
}

func UpdateAgent(agent models.Agent) error {
	platform.Logger.Info("Adding agent")
	// Register Agent
	err := repositories.UpdateAgent(agent)
	if err != nil {
		platform.Logger.Error("Unable to register normal agent in consul: ", zap.Error(err))
		return err
	}
	// Register Metrics
	//err = repositories.UpdateChaosAgent(agent, consulAgentMetricsName, agent.MetricsPort)
	//if err != nil {
	//	platform.Logger.Error("Unable to register metrics agent in consul: ", zap.Error(err))
	//	return err
	//}

	platform.Logger.Info("Added agent successfully")
	return nil
}

func DeleteAgent(agent models.Agent) error {
	platform.Logger.Info("Deleting agent")

	err := repositories.DeleteAgent(agent.Id)
	if err != nil {
		platform.Logger.Error("Unable to delete normal agent in consul: ", zap.Error(err))
		return err
	}
	//
	//err = repositories.DeleteChaosAgent(agent, consulAgentMetricsName)
	//if err != nil {
	//	platform.Logger.Error("Unable to delete metric agent in consul: ", zap.Error(err))
	//	return err
	//}

	platform.Logger.Debug("Agent deleted")
	return nil

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
		platform.Logger.Error("Unable to get Test Group: ", zap.Error(err))
		return group, err
	}

	return group, err
}

func CreateTestGroup(group models.TestGroup) error {
	err := repositories.AddTestGroup(group)
	if err != nil {
		platform.Logger.Error("Unable to create new TestGroup", zap.Error(err))
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
		platform.Logger.Error("Error verifying if the test group exists when adding test collection", zap.Error(err))
		return err
	}

	if result == false {
		platform.Logger.Warn("Parent Test Group must exist when adding a test collection")
		return ErrParentTestGroupDoesNotExist
	}

	err = repositories.AddTestCollection(test)

	return err
}

func UpdateTestCollection(test models.TestCollection) error {
	result, err := repositories.DoesTestGroupExist(test.GroupId)
	if err != nil {
		platform.Logger.Error("Error verifying if the test group exists when adding test collection", zap.Error(err))
		return err
	}

	if result == false {
		platform.Logger.Warn("Parent Test Group must exist when adding a test collection")
		return ErrParentTestGroupDoesNotExist
	}

	err = repositories.UpdateTestCollection(test)

	return err
}

func DeleteTestCollection(id string) error {
	err := repositories.DeleteTestCollection(id)
	return err
}
