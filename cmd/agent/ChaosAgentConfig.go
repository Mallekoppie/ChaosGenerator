package main

type ChaosAgentConfig struct {
	Port          string `json:"port"`
	MetricsPort   string `json:"metricsPort"`
	MasterAddress string `json:"masterAddress"`
	Version       string `json:"version"`
}
