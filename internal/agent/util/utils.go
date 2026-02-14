package util

import (
	"crypto/tls"
	"log"

	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	cpu "github.com/shirou/gopsutil/cpu"
)

var (
	httpClient       *http.Client
	DefaultTLSConfig = &tls.Config{InsecureSkipVerify: true}
)

const (
	MaxIdleConnections int = 20
	RequestTimeout     int = 30
)

// init HTTPClient
func init() {
	httpClient = createHTTPClient()
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

func MakeHttpCall(endpoint string, method string, headers map[string]string, requestBody string) (responseCode int, body string, err error) {
	var req *http.Request

	if len(requestBody) > 0 {
		requestBodyData := bytes.NewBufferString(requestBody)

		req, err = http.NewRequest(method, endpoint, requestBodyData)
	} else {
		req, err = http.NewRequest(method, endpoint, nil)
	}

	if err != nil {
		log.Fatalf("Error Occured. %+v", err)
	}

	// Add the headers to the request
	for headerKey, headerValue := range headers {
		req.Header.Add(headerKey, headerValue)
	}

	// use httpClient to send request
	response, err := httpClient.Do(req)
	if err != nil && response == nil {
		log.Fatalf("Error sending request to API endpoint. %+v", err)
		return 0, "", err
	}
	// Close the connection to reuse it
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Couldn't read response body. %+v", err)

		return 0, "", err
	}

	responseString := string(responseBody[:len(responseBody)])

	return response.StatusCode, responseString, nil
}

func GetCPUStatus() float64 {
	var cpuUsage float64
	data, cpuStatusErr := cpu.Times(true)
	var valueRetrieved bool

	for maxRetries := 0; maxRetries < 5; maxRetries++ {

		for i := range data {
			if data[i].CPU == "_Total" {
				cpuUsage = data[i].User
				valueRetrieved = true
				break
			}
		}

		if cpuStatusErr != nil || len(data) < 1 {
			if cpuStatusErr != nil {
				log.Println("Error retrieving CPU stats:", cpuStatusErr)
			}

			data, cpuStatusErr = cpu.Times(true)
			continue
		}
	}

	if valueRetrieved == false {
		cpuUsage = -1
	}

	return cpuUsage
}
