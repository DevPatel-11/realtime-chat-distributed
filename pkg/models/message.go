package models

import "time"

// MessageStatus represents the delivery state of a message
type MessageStatus string

const (
	StatusSent      MessageStatus = "sent"
	StatusDelivered MessageStatus = "delivered"
	StatusRead      MessageStatus = "read"
)

// Message represents a chat message
type Message struct {
	ID        string        `json:"id" bson:"_id"`
	SenderID  string        `json:"sender_id" bson:"sender_id"`
	ReceiverID string       `json:"receiver_id" bson:"receiver_id"`
	Content   string        `json:"content" bson:"content"`
	Timestamp time.Time     `json:"timestamp" bson:"timestamp"`
	Status    MessageStatus `json:"status" bson:"status"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at"`
}
