package repositories

import (
	"mallekoppie/ChaosGenerator/internal/master/models"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"

	"encoding/json"
)

const (
	bucketAgent string = "agent"
)

func init() {

}

func AddAgent(agent models.Agent) error {
	err := platform.Database.BoltDb.SaveObject(bucketAgent, agent.Id, agent)
	if err != nil {
		platform.Log.Error("Error saving agent", zap.Error(err))
		return err
	}

	return nil
}

func DeleteAgent(id string) error {
	err := platform.Database.BoltDb.RemoveObject(bucketAgent, id)
	if err != nil {
		platform.Log.Error("Error removing agent", zap.Error(err))
		return err
	}

	return nil
}

func DeleteAllAgents() error {
	err := platform.Database.BoltDb.RemoveBucket(bucketAgent)
	if err != nil {
		platform.Log.Error("Error removing all agents", zap.Error(err))
		return err
	}

	return nil
}

func UpdateAgent(agent models.Agent) error {
	err := platform.Database.BoltDb.SaveObject(bucketAgent, agent.Id, agent)
	if err != nil {
		platform.Log.Error("Error Updating agent", zap.Error(err))
		return err
	}

	return nil
}

func GetAllAgents() (agents []models.Agent, err error) {
	agents = make([]models.Agent, 0)

	results, err := platform.Database.BoltDb.ReadAllObjects(bucketAgent)
	if err != nil && err == platform.ErrNoEntryFoundInDB {
		return agents, nil
	}
	if err != nil {
		platform.Log.Error("Error retrieving agents", zap.Error(err))
		return nil, err
	}

	for _, v := range results {
		agent := models.Agent{}
		err = json.Unmarshal([]byte(v), &agent)
		if err != nil {
			platform.Log.Error("Error unmarshalling agent", zap.Error(err))
			return nil, err
		}

		agents = append(agents, agent)
	}

	return agents, nil
}
