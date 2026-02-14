package repositories

import (
	"mallekoppie/ChaosGenerator/internal/master/models"
	"testing"

	"github.com/google/uuid"
)

const (
	ServiceName string = "ChaosAgent"
)

func TestConsulAgentRegistration(t *testing.T) {
	agent := models.Agent{
		Id:      uuid.New().String(),
		Host:    "unittesthost",
		Port:    1100,
		Enabled: true,
	}

	err := UpdateChaosAgent(agent, ServiceName, 11000)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestConsulAgentDelete(t *testing.T) {
	agent := models.Agent{
		Id:      uuid.New().String(),
		Host:    "unittesthostdelete",
		Port:    1100,
		Enabled: true,
	}

	err := UpdateChaosAgent(agent, ServiceName, 11000)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = DeleteChaosAgent(agent, ServiceName)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestConsulAgentDisable(t *testing.T) {
	agent := models.Agent{
		Id:      uuid.New().String(),
		Host:    "unittestToDisable",
		Port:    1100,
		Enabled: true,
	}

	err := UpdateChaosAgent(agent, ServiceName, 11000)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	agent.Enabled = false

	err = UpdateChaosAgent(agent, ServiceName, 11000)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestGetAllAgents(t *testing.T) {
	agents, err := GetAllChaosAgents(ServiceName)
	if err != nil {
		t.Fatal("Unable to retrieve all agents: ", err.Error())
		t.Fail()
	}

	for i := range agents {
		t.Log("Agent returned: ", agents[i])
	}
}
