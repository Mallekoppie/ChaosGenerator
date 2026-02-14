package util

import (
	//	"log"
	"net/http"
	"testing"
	//"time"
	json "encoding/json"
	io "io/ioutil"
)

/*
// Function is slow. Commenting out while testing http functions
func TestGetCPU(t *testing.T) {

	result := GetCPUStatus()

	if result == 0 {
		t.Log("Result: ", result)
		t.Fail()
	}
}
*/

type Route struct {
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var (
	HeaderTestChan chan bool
	BodyTestChan   chan bool
	Routes         map[int]Route
	srv            http.Server
)

const (
	BasePath string = "http://localhost:9999"
)

func init() {
	HeaderTestChan = make(chan bool, 10)
	BodyTestChan = make(chan bool, 10)
	Routes = make(map[int]Route)
	srv = http.Server{Addr: ":9999"}
	Routes[0] = Route{Pattern: "/TestHttpCallForBasicGet", HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}}
	Routes[1] = Route{Pattern: "/TestHttpCallForToSendHeaders", HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
		first := false
		second := false
		for header := range r.Header {

			if header == "First" {
				if r.Header.Get(header) == "value" {
					first = true
				}
			} else if header == "Second" {
				if r.Header.Get(header) == "secondvalue" {
					second = true
				}
			}
		}

		if first == true && second == true {
			HeaderTestChan <- true
		} else {
			HeaderTestChan <- false
		}

		w.WriteHeader(http.StatusOK)
	}}
	Routes[2] = Route{Pattern: "/TestHttpCallWithBodyAndPostMethod", HandlerFunc: func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		data, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		if len(data) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		testBody := TestRequestForBodyTest{}

		marshalErr := json.Unmarshal(data, &testBody)

		if marshalErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if testBody.Name == "test" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}}

	Routes[3] = Route{Pattern: "/TestGetBody", HandlerFunc: func(w http.ResponseWriter, r *http.Request) {

		testBody := TestRequestForBodyTest{}
		testBody.Name = "test"

		data, _ := json.Marshal(testBody)

		w.WriteHeader(http.StatusOK)
		w.Write(data)

	}}

	go MakeLittleServerToCall()
}

func TestGetBody(t *testing.T) {
	headers := make(map[string]string)

	responseCode, responseBody, _ := MakeHttpCall(BasePath+"/TestGetBody", http.MethodPost, headers, "")

	if responseCode != http.StatusOK {
		t.Log("Incorrect response code received")
		t.Fail()
	}

	testBody := TestRequestForBodyTest{}

	json.Unmarshal([]byte(responseBody), &testBody)

	if testBody.Name != "test" {
		t.Log("Body data not correct")
		t.Fail()
	}

}

type TestRequestForBodyTest struct {
	Name    string
	Surname string
}

func TestHttpCallWithBodyAndPostMethod(t *testing.T) {
	headers := make(map[string]string)
	testBody := TestRequestForBodyTest{}
	testBody.Name = "test"
	data, _ := json.Marshal(testBody)

	responseCode, _, _ := MakeHttpCall(BasePath+"/TestHttpCallWithBodyAndPostMethod", http.MethodPost, headers, string(data))

	if responseCode != http.StatusOK {
		t.Log("Incorrect response code received")
		t.Fail()
	}
}

func MakeLittleServerToCall() {

	myMux := http.NewServeMux()

	for r := range Routes {
		myMux.HandleFunc(Routes[r].Pattern, Routes[r].HandlerFunc)
	}

	srv.Handler = myMux

	srv.ListenAndServe()
}

func TestHttpCallForBasicGet(t *testing.T) {

	headers := make(map[string]string)
	responseCode, _, _ := MakeHttpCall(BasePath+"/TestHttpCallForBasicGet", http.MethodGet, headers, "")

	if responseCode != http.StatusOK {
		t.Log("Incorrect response code received")
		t.Fail()
	}

	t.Log("Received correct response code")
}

func TestHttpCallForToSendHeaders(t *testing.T) {

	headers := make(map[string]string)
	headers["first"] = "value"
	headers["second"] = "secondvalue"

	responseCode, _, _ := MakeHttpCall(BasePath+"/TestHttpCallForToSendHeaders", http.MethodGet, headers, "")

	if responseCode != http.StatusOK {
		t.Log("Incorrect response code received")
		t.Fail()
	}

	t.Log("Waiting for Channel")
	correctheadersReceived := <-HeaderTestChan

	if correctheadersReceived != true {
		t.Log("Incorrect headers received")
		t.Fail()
	}
}

/*
func TestExternalTest(t *testing.T) {
	headers := make(map[string]string)
	headers["first"] = "value"
	headers["test"] = "secondvalue"

	responseCode, _, _ := MakeHttpCall("http://localhost:80/ConnectionTest", http.MethodGet, headers, "")

	if responseCode != http.StatusOK {
		t.Log("Incorrect response code received")
		t.Fail()
	}
}
*/

// This is done so that the server gracefully shuts down
func TestKillServer(t *testing.T) {
	srv.Close()
}
