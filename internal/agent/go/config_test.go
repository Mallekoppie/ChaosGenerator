package swagger

import (
	base64 "encoding/base64"
	json "encoding/json"
	"net/http"
	"testing"

	pb "mallekoppie/ChaosGenerator/contract"
)

func init() {
	config := pb.TestCollection{Name: "test"}
	WriteTestConfiguration(config)
}

func TestReadConfig(t *testing.T) {
	test, err := ReadTestConfiguration("test")

	if err != nil {
		t.Logf("Failed to read config: %v", err)
		t.Fail()
	}

	if len(test.Name) > 0 {
		t.Logf("Test Collection retrieved: %v", test.Name)
	}

}

func TestWriteConfig(t *testing.T) {
	config := pb.TestCollection{Name: "test"}

	WriteTestConfiguration(config)
}

type TestWithBase64Body struct {
	Name string
}

func TestWriteConfigWithBase64Body(t *testing.T) {
	config := pb.TestCollection{Name: "testBase64Body"}
	body := TestWithBase64Body{Name: "TestValue"}

	data, _ := json.Marshal(body)

	encodedData := base64.StdEncoding.EncodeToString(data)

	config.Tests = []*pb.Test{{
		Name:         "BasicGet",
		Method:       "GET",
		Body:         encodedData,
		Url:          "http://localhost:9001/TestTestRunnerFirstBasic",
		ResponseCode: http.StatusOK,
	},
	}

	WriteTestConfiguration(config)

	tests, err := ReadTestConfiguration("testBase64Body")

	if err != nil {
		t.Log("Error during config read", err)
		t.Fail()
	}

	retrievedConfig := TestWithBase64Body{}

	json.Unmarshal([]byte(tests.Tests[0].Body), &retrievedConfig)

	if retrievedConfig.Name != "TestValue" {
		t.Log("Retrieved body is incorrect: ", retrievedConfig.Name)
		t.Fail()
	}

}

func TestCreateTestRunnerConfig(t *testing.T) {
	config := pb.TestCollection{Name: "TestRunnerFirst"}
	config.Tests = []*pb.Test{{
		Name:         "BasicGet",
		Method:       "GET",
		Body:         "",
		Url:          "http://localhost:9001/TestTestRunnerFirstBasic",
		ResponseCode: http.StatusOK,
	},
	}

	WriteTestConfiguration(config)
}
