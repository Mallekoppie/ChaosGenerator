package models

type ConsulRequest struct {
	Service struct {
		Service string `json:"service"`
		Port    int    `json:"port"`
	} `json:"service"`
	Node     string `json:"node"`
	Address  string `json:"address"`
	NodeMeta struct {
		Enabled string `json:"enabled"`
		Id      string `json:"id"`
	} `json:"NodeMeta"`
}
