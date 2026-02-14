package models

type Agent struct {
	Id          string `json:"id"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	MetricsPort int    `json:"metricsPort"`
	Enabled     bool   `json:"enabled"`
	Status      string `json:"status"`
}
