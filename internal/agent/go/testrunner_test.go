package swagger

import (
	"net/http"
	"testing"
	"time"
	//json "encoding/json"
	//io "io/ioutil"
)

/*
func TestGetTestStatus(t *testing.T) {
	status := CoreGetTestStatus()

	t.Log(status.Cpu)
	t.Log(status.ExecutionTime)
	t.Log(status.RequestsExecuted)
	t.Log(status.SimulatedUsers)
	t.Log(status.TestCollectionName)
	t.Log(status.TransactionsPerSecond)
}

*/

type TestRoute struct {
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var (
	TestCountChen chan int
	TestRoutes    map[int]TestRoute
	srv           http.Server
)

const (
	BasePath string = "http://localhost:9001"
)

func MakeLittleServerToCall() {

	myMux := http.NewServeMux()

	for r := range TestRoutes {
		myMux.HandleFunc(TestRoutes[r].Pattern, TestRoutes[r].HandlerFunc)
	}

	srv.Handler = myMux

	srv.ListenAndServe()
}

var TestCount int

func init() {
	TestRoutes = make(map[int]TestRoute)
	srv = http.Server{Addr: ":9001"}
	TestRoutes[0] = TestRoute{Pattern: "/TestTestRunnerFirstBasic", HandlerFunc: func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)
		TestCount++
		time.Sleep(time.Millisecond * 200)
	}}

	go MakeLittleServerToCall()
}

func TestTestRunnerFirstBasic(t *testing.T) {
	CoreRunTest("TestRunnerFirst", 1)

	time.Sleep(time.Second * 2)

	CoreStopTest()

	//t.Log("Tests completed: ", TestCount)

}

func TestTestRunnerGetTPS(t *testing.T) {
	t.Log("Tests completed: ", TestCount)
	CoreRunTest("TestRunnerFirst", 1)

	time.Sleep(time.Second * 11)

	status := CoreGetTestStatus()
	t.Logf("AverageExecutionTime: \t %v, ExecutionTime: \t %v, RequestsExecuted: \t %v, SimulatedUsers: \t %v, TransactionsPerSecond: \t %v", status.AverageExecutionTime, status.ExecutionTime, status.RequestsExecuted, status.SimulatedUsers, status.TransactionsPerSecond)

	CoreStopTest()
	t.Log("Tests completed: ", TestCount)
}

func TestKillServer(t *testing.T) {
	srv.Close()
}
