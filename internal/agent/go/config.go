package swagger

import (
	base64 "encoding/base64"
	json "encoding/json"
	io "io/ioutil"
	"log"
	"os"

	"github.com/tkanos/gonfig"

	pb "mallekoppie/ChaosGenerator/contract"
)

func ClearTestsDirectory() error {
	err := os.RemoveAll("tests")
	if err != nil {
		log.Println("Unable to remove tests directory: ", err.Error())
		return err
	}

	return nil
}

func WriteTestConfiguration(config pb.TestCollection) error {
	data, err := json.Marshal(config)

	if err != nil {
		log.Printf("Failed to marshall config: %v", err)
		return err
	}

	os.MkdirAll("tests", 0755)

	name := "tests/" + config.Name + ".json"

	err = io.WriteFile(name, data, os.ModeExclusive)

	if err != nil {
		log.Printf("Failed to marshall config: %v", err)
		return err
	}

	return nil
}

// The body of the requests for each test must be base64 encoded.
// We don't do anything when writing but when reading we must decode it
func ReadTestConfiguration(name string) (pb.TestCollection, error) {
	configuration := pb.TestCollection{}
	err := gonfig.GetConf("tests/"+name+".json", &configuration)

	if err != nil {
		log.Printf("Error reading config: %v", err)
		return configuration, err
	}

	//Decode bodies
	for i := 0; i < len(configuration.Tests); i++ {
		if len(configuration.Tests[i].Body) > 0 {
			data, err := base64.StdEncoding.DecodeString(configuration.Tests[i].Body)

			if err != nil {
				log.Println("Error base64 decoding request body: ", err)
			}

			configuration.Tests[i].Body = string(data)
		}
	}

	return configuration, nil
}
