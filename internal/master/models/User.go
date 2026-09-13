package models

import "time"

// User represents a local user that can authenticate against the master.
// PasswordHash is persisted (BoltDB marshals this struct as JSON) but is never
// exposed through the gRPC contract.
type User struct {
	Id           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"passwordHash"`
	CreatedAt    time.Time `json:"createdAt"`
}
