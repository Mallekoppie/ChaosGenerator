package models

import "time"

// Target represents a test target service
type Target struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	Address                string    `json:"address"`
	ServiceType            string    `json:"serviceType"` // "http/1.1", "http/2", "grpc"
	Protocol               string    `json:"protocol"`    // "http", "https", "grpc"
	ConnectionReuseEnabled bool      `json:"connectionReuseEnabled"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}
