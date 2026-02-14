package repositories

import (
	"mallekoppie/ChaosGenerator/internal/master/models"
	"testing"
)

func TestWriteAndReadConfig(t *testing.T) {
	config := models.ServiceConfig{
		GrafanaUrl:    "http://localhost:3000",
		PrometheusUrl: "http://localhost:9090",
		MongoDBHost:   "localhost",
		MongoDBPort:   27017,
		MongoDBName:   "UnitTest",
		ConsulUrl:     "http://127.0.0.1:8500",
		ConsulToken:   "none",
	}

	err := WriteConfig(config)
	if err != nil {
		t.Log("Unable to write config: ", err.Error())
		t.Fail()
		return
	}

	readConfig, err := GetConfig()
	if err != nil {
		t.Log("Unable to read config: ", err.Error())
		t.Fail()
		return
	}

	if readConfig.MongoDBPort != 27017 {
		t.Log("Incorrect value read back")
		t.Fail()
		return
	}

	t.Log("Correct value returned when reading config")
}
