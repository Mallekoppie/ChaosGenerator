package repositories

import (
	"mallekoppie/ChaosGenerator/internal/master/models"

	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/tkanos/gonfig"
)

const (
	ConfigFileName string = "ChaosMasterConfig.json"
)

func GetConfig() (models.ServiceConfig, error) {
	configuration := models.ServiceConfig{}
	err := gonfig.GetConf(ConfigFileName, &configuration)

	if err != nil {
		log.Printf("Error reading config file %v: %v", ConfigFileName, err.Error())
		return configuration, err
	}

	return configuration, nil
}

func WriteConfig(config models.ServiceConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		log.Println("Unable to marshall config to json: ", err.Error())
		return err
	}

	ioutil.WriteFile(ConfigFileName, data, os.ModeExclusive)
	if err != nil {
		log.Println("Unable to write config to file: ", err.Error())
		return err
	}

	return nil
}
