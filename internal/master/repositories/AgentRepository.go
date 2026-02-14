package repositories

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mallekoppie/ChaosGenerator/internal/master/models"
	"net/http"
	"strconv"
	"time"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
)

var (
	httpClient                     *http.Client
	DefaultTLSConfig               = &tls.Config{InsecureSkipVerify: true}
	ConsulUrl                      string
	ConsulToken                    string
	ErrConsulPortIncorrect         = errors.New("consul port incorrect")
	ErrConsulHostIncorrect         = errors.New("consul host incorrect")
	ErrConsulResponseCodeIncorrect = errors.New("consul returned incorrect response code")
)

const (
	MaxIdleConnections int = 20
	RequestTimeout     int = 30
)

var (
	serviceConfig models.ServiceConfig
)

// init HTTPClient
func init() {
	httpClient = createHTTPClient()

	serviceConfig, _ = GetConfig()
	ConsulUrl = serviceConfig.ConsulUrl
	ConsulToken = serviceConfig.ConsulToken
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: MaxIdleConnections,
			TLSClientConfig:     DefaultTLSConfig,
		},
		Timeout: time.Duration(RequestTimeout) * time.Second,
	}

	return client
}

func createConsulRequest(port int, host string, enabled bool, serviceName string, id string) (request models.ConsulRequest, err error) {
	consulRequest := models.ConsulRequest{}

	if port < 1024 || port > 65200 {
		log.Println("Bad port value for consul request: ", port)
		platform.Log.Error("Bad port value for consul request", zap.Int("port", port))
		return consulRequest, ErrConsulPortIncorrect
	}

	if len(host) < 1 {
		platform.Log.Error("Bad host value for consul request", zap.String("host", host))
		return consulRequest, ErrConsulHostIncorrect
	}

	consulRequest.Service.Service = serviceName
	consulRequest.Service.Port = port
	consulRequest.Node = fmt.Sprintf("%v:%v:%v", serviceName, host, port)
	consulRequest.NodeMeta.Enabled = strconv.FormatBool(enabled)
	consulRequest.NodeMeta.Id = id
	consulRequest.Address = host

	return consulRequest, nil
}

func UpdateChaosAgent(agent models.Agent, serviceName string, port int) error {

	requestObject, err := createConsulRequest(port, agent.Host, agent.Enabled, serviceName, agent.Id)
	if err != nil {
		return err
	}

	data, err := json.Marshal(requestObject)
	if err != nil {
		platform.Log.Error("Unable to marchall agent to json for update: ", zap.Error(err))
		return err
	}

	request, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v/v1/catalog/register", ConsulUrl), bytes.NewBuffer(data))
	if err != nil {
		platform.Log.Error("Unable to create new HTTP request for Consul update: ", zap.Error(err))
		return err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		platform.Log.Error("Error while sending request to Consul: ", zap.Error(err))
		return err
	}

	if response.StatusCode != http.StatusOK {
		platform.Log.Error("Incorrect response code from consul for agent registration: ", zap.Int("status_code", response.StatusCode))
		return ErrConsulResponseCodeIncorrect
	}

	return nil
}

func DeleteChaosAgent(agent models.Agent, serviceName string) error {
	requestObject, err := createConsulRequest(agent.Port, agent.Host, agent.Enabled, serviceName, agent.Id)
	if err != nil {
		return err
	}

	data, err := json.Marshal(requestObject)
	if err != nil {
		platform.Log.Error("Unable to marchall agent to json for update: ", zap.Error(err))
		return err
	}

	request, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%v/v1/catalog/deregister", ConsulUrl), bytes.NewBuffer(data))
	if err != nil {
		platform.Log.Error("Unable to create new HTTP request for Consul update: ", zap.Error(err))
		return err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		platform.Log.Error("Error while sending request to Consul: ", zap.Error(err))
		return err
	}

	if response.StatusCode != http.StatusOK {
		platform.Log.Error("Incorrect response code from consul for agent registration: ", zap.Int("status_code", response.StatusCode))
		return ErrConsulResponseCodeIncorrect
	}

	return nil
}

func GetAllChaosAgents(serviceName string) (agents []models.Agent, err error) {

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/v1/catalog/service/%v", ConsulUrl, serviceName), nil)
	if err != nil {
		platform.Log.Error("Error creating request to retrieve all consul agents: ", zap.Error(err))
		return nil, err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		platform.Log.Error("Unable to make call to consul: ", zap.Error(err))
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		platform.Log.Error("Consul returned incorrect response code. Expected 200 but received ", zap.Int("status_code", response.StatusCode))
		return nil, ErrConsulResponseCodeIncorrect
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		platform.Log.Error("Error reading response body: ", zap.Error(err))
		return nil, err
	}
	consulAgentResponse := models.ConsulAgentResponse{}
	err = json.Unmarshal(data, &consulAgentResponse)
	if err != nil {
		platform.Log.Error("Error unmarshalling consul agents: ", zap.Error(err))
		return nil, err
	}

	agents = make([]models.Agent, 0)
	for i := range consulAgentResponse {

		agent := models.Agent{
			Id:   consulAgentResponse[i].NodeMeta.Id,
			Host: consulAgentResponse[i].Address,
			Port: consulAgentResponse[i].ServicePort,
		}
		enabled, err := strconv.ParseBool(consulAgentResponse[i].NodeMeta.Enabled)
		if err != nil {
			platform.Log.Error("unable to parse agent enabled status: ", zap.Error(err))
			enabled = false
		} else {
			agent.Enabled = enabled
		}

		agents = append(agents, agent)
	}

	return agents, nil
}
