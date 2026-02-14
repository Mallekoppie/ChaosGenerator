package manager

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/tkanos/gonfig"

	pb "mallekoppie/ChaosGenerator/contract"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"mallekoppie/ChaosGenerator/internal/master/repositories"
	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
)

var (
	agents                 []ChaosAgent
	ErrorNoAgentWithThatId = errors.New("No agent with that Id")
)

const (
	consulAgentServiceName string = "ChaosAgent"
)

type ChaosAgent struct {
	Id     string
	Name   string `json:"name"`
	Url    string `json:"url"`
	Client pb.ChaosAgentClient
	Ctx    context.Context
}

func init() {
	initializeAgents()
}

func getAgents() []ChaosAgent {
	if len(agents) < 1 {
		platform.Logger.Info("No agents. Re-initializing")
		initializeAgents()
	}

	return agents
}

func initializeAgents() {
	err := GetChaosMasterAgents()

	if err != nil {
		fmt.Println("Error retrieving config: ", err)
		return
	}

	platform.Logger.Info("Initializing agents")
	for i := range agents {
		err := agents[i].Init()
		if err != nil {
			platform.Logger.Error("Unable to initialize agent", zap.Error(err))
		}
	}
	count := len(agents)
	platform.Logger.Info("Config count during initialization", zap.Int("count", count))
	platform.Logger.Info("Initialized agents")
}

func GetChaosMasterAgents() error {
	agents = make([]ChaosAgent, 0)
	consulAgents, err := repositories.GetAllAgents(consulAgentServiceName)
	if err != nil {
		platform.Logger.Error("Unable to get agent configuration", zap.Error(err))
		return err
	}

	for index := range consulAgents {
		agent := consulAgents[index]
		c := ChaosAgent{
			Id:   agent.Id,
			Name: agent.Host,
			Url:  fmt.Sprintf("%v:%v", agent.Host, agent.Port),
		}
		agents = append(agents, c)
	}

	return nil
}

func GetAgent(id string) (agent ChaosAgent, err error) {
	nulAgent := ChaosAgent{}

	agents := getAgents()
	number := len(agents)
	platform.Logger.Info("Number of agents returned",zap.Int("agent_number", number))

	for i := range agents {
		log.Println("inside loop")
		log.Printf("Comparing %v to %v", agents[i].Id, id)
		if agents[i].Id == id {
			platform.Logger.Debug("Agent Found")
			return agents[i], nil
		}
	}

	return nulAgent, ErrorNoAgentWithThatId
}

func GetTest(testName string) (pb.TestCollection, error) {
	configuration := pb.TestCollection{}
	err := gonfig.GetConf("./tests/"+testName+".json", &configuration)

	if err != nil {
		log.Printf("Error reading config: %v", err)
		return configuration, err
	}

	return configuration, nil
}

func (c *ChaosAgent) Init() error {
	creds, err := credentials.NewClientTLSFromFile("./chaos_agent.cer", "chaos-agent")
	if err != nil {
		log.Println("Error reading certificate: ", err.Error())
		return err
	}

	//conn, err := grpc.Dial(c.Url, grpc.WithTransportCredentials(creds))
	conn, err := grpc.Dial(c.Url, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Printf("Unable to connect to %v. Error: %v", c.Url, err.Error())
		return err
	}
	log.Println("Connection state: ", conn.GetState().String())

	c.Client = pb.NewChaosAgentClient(conn)
	c.Ctx = context.TODO()

	resp, err := c.Client.GetVersion(c.Ctx, &pb.Request{})
	if err != nil {
		log.Printf("Error during version check to %s. Error: %s", c.Url, err.Error())
		return err
	}

	log.Printf("%v is online with version %v", c.Url, resp.GetVersion())

	return nil
}

func (c *ChaosAgent) GetStatus() (pb.TestStatus, error) {

	status, err := c.Client.GetTestStatus(c.Ctx, &pb.Request{})

	if err != nil {
		//fmt.Println("Error calling service: ", err)
		return *status, err
	} else {
		return *status, nil
	}

}

func (c *ChaosAgent) IsAlive() bool {
	response, err := c.Client.IsAlive(c.Ctx, &pb.Request{})

	if err != nil {
		//fmt.Printf("Error checking if %v is alive. Error: %v", c.Name, err)
		return false
	} else if response != nil && response.Result == true {
		return true
	}

	return false

}

func (c *ChaosAgent) AddTest(test pb.TestCollection) {
	resp, err := c.Client.AddTests(c.Ctx, &test)

	if err != nil {
		fmt.Printf("Error adding test to %v . Error: %v", c.Name, err)
		return
	}

	if resp != nil && resp.Result != true {
		fmt.Printf("Error adding test for %v . Result: %v", c.Name, resp.Result)
	}
}

func (c *ChaosAgent) StartTest(testParameters pb.TestParameters) {
	resp, err := c.Client.StartTestRun(c.Ctx, &testParameters)

	if err != nil {
		fmt.Printf("Error starting test to %v . Error: %v", c.Name, err)
	}

	if resp != nil && resp.Result != true {
		fmt.Printf("Error starting test for %v . Result: %v", c.Name, resp.Result)
	}
}

func (c *ChaosAgent) UpdateTest(testParameters pb.TestParameters) {
	resp, err := c.Client.UpdateTestRun(c.Ctx, &testParameters)

	if err != nil {
		fmt.Printf("Error updating test to %v . Error: %v", c.Name, err)
	}

	if resp != nil && resp.Result != true {
		fmt.Printf("Error updating test for %v . Result: %v", c.Name, resp.Result)
	}
}

func (c *ChaosAgent) StopTest() {
	if c != nil && c.Client != nil {
		resp, err := c.Client.StopTestRun(c.Ctx, &pb.StopTestRequest{})

		if err != nil {
			fmt.Printf("Error stopping test to %v . Error: %v", c.Name, err)
		}

		if resp != nil && resp.Result != true {
			fmt.Printf("Error stopping test for %v . Result: %v", c.Name, resp.Result)
		}
	}
}

func (c *ChaosAgent) GetVersion() (pb.GetVersionResponse, error) {

	version, err := c.Client.GetVersion(c.Ctx, &pb.Request{})

	if err != nil {
		return *version, err
	} else {
		return *version, nil
	}
}

func (c *ChaosAgent) DeleteTests() {
	_, err := c.Client.DeleteTests(c.Ctx, &pb.DeleteTestsRequest{})
	if err != nil {
		fmt.Printf("Agent %v encountered error while deleting tests directory: %v", c.Url, err.Error())
		return
	}

	fmt.Println("Tests cleared on agent: ", c.Url)
}

func (c *ChaosAgent) Shutdown() {

}
