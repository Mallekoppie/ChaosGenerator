package repositories

import (
	"fmt"
	"mallekoppie/ChaosGenerator/internal/master/models"
	"testing"

	"github.com/Mallekoppie/goslow/platform"

	"github.com/google/uuid"
)

func TestFormatting(t *testing.T) {

	result := fmt.Sprintf("mongodb://%v:%v", "localhost", 27017)

	t.Log(result)

}

func createAgent() models.Agent {
	return models.Agent{
		Id:      uuid.New().String(),
		Host:    "unittesthostdelete",
		Port:    1100,
		Enabled: true,
	}
}

func TestCreateAndReadAgent(t *testing.T) {
	agent := createAgent()

	err := AddAgent(agent)
	if err != nil {
		t.Error("Failed to save agent", err.Error())
		t.Fail()
	}

	result, err := GetAllAgents()
	if err != nil {
		t.Fail()
	}

	if len(result) < 1 {
		fmt.Println("Not enough agents returned")
		t.Fail()
	}

	for _, v := range result {
		fmt.Println(v)
	}

	err = DeleteAgent(agent.Id)
	if err != nil {
		t.Error("Error deleting agent: ", err.Error())
		t.Fail()
	}

	result2, err := GetAllAgents()
	if err != nil {
		t.Error("Error getting all agents: ", err.Error())
		t.Fail()
	}

	for _, v := range result2 {
		if v.Id == agent.Id {
			t.Error("The agent is still in the DB")
			t.Fail()
		}
	}

}

func TestGetWhenThereAreNoAgents(t *testing.T) {
	err := DeleteAllAgents()
	if err != nil {
		t.Error("Error removing all agents")
		t.Fail()
	}

	_, err = GetAllAgents()
	if err != nil && err == platform.ErrNoEntryFoundInDB {

	} else if err != nil {
		t.Error("Error getting all agents")
		t.Fail()
	}
}

// func TestDeleteAllTestGroups(t *testing.T) {
// 	err := DeleteAllTestGroups()
// 	if err != nil {
// 		t.Log("Failed to delete all Test Groups: ", err.Error())
// 		t.Fail()
// 	}
// }
