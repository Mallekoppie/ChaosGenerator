package models

type ServiceConfig struct {
	GrafanaUrl      string `json:"grafana_url"`
	PrometheusUrl   string `json:"prometheus_url"`
	IdpUrl          string `json:"idp_url"`
	IdpClientId     string `json:"idp_client_id"`
	MongoDBName     string `json:"mongo_db_name"`
	MongoDBHost     string `json:"mongo_db_host"`
	MongoDBPort     int    `json:"mongo_db_port"`
	MongoDBUsername string `json:"mongo_db_username"`
	MongoDBPassword string `json:"mongo_db_password"`
	VaultEnabled    bool   `json:"vault_enabled"`
	ConsulUrl       string `json:"consul_url"`
	ConsulToken     string `json:"consul_token"`
}
