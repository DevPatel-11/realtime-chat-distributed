package contracts

import "github.com/DevPatel-11/realtime-chat-distributed/pkg/models"

// ChatService defines chat business logic operations
type ChatService interface {
	// SendMessage validates, persists, and publishes message
	SendMessage(msg *models.Message) error
	
	// GetMessages retrieves message history
	GetMessages(userID string, otherUserID string, limit int) ([]*models.Message, error)
	
	// UpdateMessageStatus updates message delivery status
	UpdateMessageStatus(messageID string, status models.MessageStatus) error
}
