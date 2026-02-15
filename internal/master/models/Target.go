package models

import "time"

// Target represents a test target service
type Target struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	Address                string    `json:"address"`
	Protocol               string    `json:"protocol"` // "http", "https", "grpc"
	ConnectionReuseEnabled bool      `json:"connectionReuseEnabled"`
	SNI                    string    `json:"sni"` // Server Name Indication for TLS requests
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}
