package models

import "time"

// UserSession represents an active user connection
type UserSession struct {
	UserID       string    `json:"user_id"`
	ConnectionID string    `json:"connection_id"`
	GatewayAddr  string    `json:"gateway_addr"`
	ConnectedAt  time.Time `json:"connected_at"`
}
