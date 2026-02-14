package models

type ConsulAgentResponse []struct {
	ID       string `json:"ID"`
	Node     string `json:"Node"`
	Address  string `json:"Address"`
	NodeMeta struct {
		Enabled string `json:"enabled"`
		Id      string `json:"id"`
	} `json:"NodeMeta"`
	ServiceName string `json:"ServiceName"`
	ServicePort int    `json:"ServicePort"`
}
