package targetgrpc

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	MetricsPort int
	CertFile    string // Path to TLS certificate file
	KeyFile     string // Path to TLS key file
}

func LoadConfig() Config {
	return Config{
		Port:        getEnvInt("CHAOS_TARGET_PORT", 9000),
		MetricsPort: getEnvInt("CHAOS_TARGET_METRICS_PORT", 9090),
		CertFile:    getEnv("CHAOS_TARGET_CERT_FILE", ""),
		KeyFile:     getEnv("CHAOS_TARGET_KEY_FILE", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// HasTLS returns true if both certificate and key files are configured
func (c *Config) HasTLS() bool {
	return c.CertFile != "" && c.KeyFile != ""
}
