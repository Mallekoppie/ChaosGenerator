package swagger

import (
	"errors"
	util "mallekoppie/ChaosGenerator/internal/agent/util"

	"log"
	"net/http"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"

	pb "mallekoppie/ChaosGenerator/internal/contracts"
)

var (
	ExecutionTimeNanosecond int64
	RequestsExecuted        int64
	SimulatedUsers          int32
	ErrorCount              int32
	TestCollectionName      string
	IsTestRunning           bool
	SampleIntervalSecond    int32 = 10
	TestStatisticsChan      chan TestStatistics
	attacker                *vegeta.Attacker
)

func init() {
	TestStatisticsChan = make(chan TestStatistics, 2000)
	go MonitorAndUpdateStatistics()
}

type TestStatistics struct {
	RequestsExecuted             int64
	TotalExecutionTimeNanosecond int64
	ErrorCount                   int32
}

func CoreGetTestStatus() pb.TestStatus {
	testStatus := pb.TestStatus{}

	testStatus.Cpu = util.GetCPUStatus() //Slow
	if ExecutionTimeNanosecond > 0 {
		testStatus.ExecutionTime = ExecutionTimeNanosecond / 1000000000
	}
	testStatus.RequestsExecuted = RequestsExecuted
	testStatus.SimulatedUsers = SimulatedUsers
	if RequestsExecuted > 0 && SimulatedUsers > 0 {
		testStatus.AverageExecutionTime = ExecutionTimeNanosecond / RequestsExecuted / int64(SimulatedUsers) / 1000000
	}

	if testStatus.ExecutionTime > 0 {
		testStatus.TransactionsPerSecond = RequestsExecuted / testStatus.ExecutionTime
		testStatus.TransactionsPerSecond = testStatus.TransactionsPerSecond * int64(SimulatedUsers)
	}
	testStatus.TestCollectionName = TestCollectionName

	testStatus.ErrorsRaised = ErrorCount

	if ErrorCount > 0 && testStatus.ExecutionTime > 0 {
		testStatus.ErrorsPerSecond = int64(ErrorCount) / testStatus.ExecutionTime
	}

	return testStatus
}

func CoreStopTest() {
	if IsTestRunning == true {
		attacker.Stop()

		SimulatedUsers = 0
		IsTestRunning = false
	}
}

func CoreRunTest(testName string, simulatedUsersInput int) (bool, error) {

	if IsTestRunning == true {
		return IsTestRunning, errors.New("Test is already running")
	}

	if IsTestRunning == false {
		IsTestRunning = true
		SimulatedUsers = 0
		ExecutionTimeNanosecond = 0
		RequestsExecuted = 0
		ErrorCount = 0
		TestCollectionName = testName
		go StartVegeta(testName, simulatedUsersInput)
	}

	return IsTestRunning, nil
}

func StartVegeta(testName string, simulatedUsersInput int) {
	//SimulatedUsers = int32(simulatedUsersInput)
	rate := vegeta.Rate{Freq: simulatedUsersInput, Per: time.Second}

	duration := 0 * time.Second
	tests, _ := ReadTestConfiguration(testName)

	targets := make([]vegeta.Target, 0)

	for testIndex := range tests.Tests {
		headers := http.Header{}
		if len(tests.Tests[testIndex].Headers) > 0 {
			for headerIndex := range tests.Tests[testIndex].Headers {
				headers.Add(tests.Tests[testIndex].Headers[headerIndex].Name, tests.Tests[testIndex].Headers[headerIndex].Value)
			}
		}

		targets = append(targets, vegeta.Target{
			Method: tests.Tests[testIndex].Method,
			URL:    tests.Tests[testIndex].Url,
			Body:   []byte(tests.Tests[testIndex].Body),
			Header: headers,
		})
	}

	targeter := vegeta.NewStaticTargeter(targets...)

	attacker = vegeta.NewAttacker()

	var metrics vegeta.Metrics
	IsTestRunning = true
	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		metrics.Add(res)

		var executed int64
		var errors int32

		if len(res.Error) > 0 {
			errors = 1

			log.Println("Error: ", res.Error)
		} else {
			executed = 1
		}

		testStats := TestStatistics{
			RequestsExecuted:             executed,
			TotalExecutionTimeNanosecond: res.Latency.Nanoseconds(),
			ErrorCount:                   errors,
		}

		TestStatisticsChan <- testStats
	}
	metrics.Close()
}

func CoreUpdateTest(simulatedUsersInput int32) error {
	if IsTestRunning == false {
		return errors.New("No Test is running")
	}

	if simulatedUsersInput < SimulatedUsers-1 {
		return nil
	}

	testCollection, configError := ReadTestConfiguration(TestCollectionName)

	if configError != nil {
		return errors.New("Could not retrieve test config")
	}

	for SimulatedUsers < simulatedUsersInput {
		go RunTest(testCollection)
		SimulatedUsers++
	}

	return nil
}

func MonitorAndUpdateStatistics() {
	lastResetTime := time.Now()
	for true {
		testStats := <-TestStatisticsChan

		RequestsExecuted = RequestsExecuted + testStats.RequestsExecuted
		ExecutionTimeNanosecond = ExecutionTimeNanosecond + testStats.TotalExecutionTimeNanosecond
		ErrorCount = ErrorCount + testStats.ErrorCount

		timeSince := time.Since(lastResetTime)

		if timeSince.Seconds() > 10 {
			RequestsExecuted = 0
			ExecutionTimeNanosecond = 0
			ErrorCount = 0
			lastResetTime = time.Now()
		}

	}
}

func RunTest(config pb.TestCollection) {

	for IsTestRunning == true {

		for testIndex := range config.Tests {
			var testErrors int32
			var testTimeNanosecond, testsCompleted int64
			if IsTestRunning == false {
				break
			}

			item := config.Tests[testIndex]
			headers := make(map[string]string)

			if len(item.Headers) > 0 {
				for h := range item.Headers {
					headers[item.Headers[h].Name] = item.Headers[h].Value
				}
			}

			startTime := time.Now()
			responseCode, responseBody, err := util.MakeHttpCall(item.Url, item.Method, headers, item.Body)
			result := time.Since(startTime)
			testTimeNanosecond = result.Nanoseconds()
			testsCompleted++

			if err != nil || (int32(responseCode)) != item.ResponseCode {
				testErrors++
				log.Printf("Error. Expected Code: %v but received: %v", item.ResponseCode, responseCode)
			} else if len(item.ResponseBody) > 0 && responseBody != item.ResponseBody {
				testErrors++
				log.Printf("Error. Expected body: %v but received: %v", item.ResponseBody, responseBody)
			}

			testStats := TestStatistics{
				RequestsExecuted:             testsCompleted,
				TotalExecutionTimeNanosecond: testTimeNanosecond,
				ErrorCount:                   testErrors,
			}

			TestStatisticsChan <- testStats
		}
	}
}
