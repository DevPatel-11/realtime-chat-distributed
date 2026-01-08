package contracts

import "github.com/DevPatel-11/realtime-chat-distributed/pkg/models"

// GatewayService defines WebSocket gateway operations
type GatewayService interface {
	// HandleConnection manages WebSocket connection lifecycle
	HandleConnection(userID string, connID string) error
	
	// SendMessage delivers a message to connected client
	SendMessage(userID string, msg *models.Message) error
	
	// CloseConnection gracefully closes a connection
	CloseConnection(userID string) error
}
