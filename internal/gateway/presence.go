package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// PresenceStore manages user presence in Redis
type PresenceStore interface {
	SetUserSession(ctx context.Context, userID string, connID string) error
	GetUserSession(ctx context.Context, userID string) (*SessionData, error)
	RemoveUserSession(ctx context.Context, userID string) error
	IsUserOnline(ctx context.Context, userID string) (bool, error)
}

// SessionData represents session information in Redis
type SessionData struct {
	ConnectionID string    `json:"connection_id"`
	GatewayAddr  string    `json:"gateway_addr"`
	ConnectedAt  time.Time `json:"connected_at"`
}

// RedisPresenceStore implements PresenceStore using Redis
type RedisPresenceStore struct {
	client      *redis.Client
	gatewayAddr string
	ttl         time.Duration
}

// NewRedisPresenceStore creates a new Redis presence store
func NewRedisPresenceStore(client *redis.Client, gatewayAddr string) *RedisPresenceStore {
	return &RedisPresenceStore{
		client:      client,
		gatewayAddr: gatewayAddr,
		ttl:         30 * time.Minute,
	}
}

// SetUserSession stores user session in Redis
func (s *RedisPresenceStore) SetUserSession(ctx context.Context, userID string, connID string) error {
	key := fmt.Sprintf("session:%s", userID)
	
	session := &SessionData{
		ConnectionID: connID,
		GatewayAddr:  s.gatewayAddr,
		ConnectedAt:  time.Now(),
	}
	
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	err = s.client.Set(ctx, key, data, s.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set session in redis: %w", err)
	}
	
	return nil
}

// GetUserSession retrieves user session from Redis
func (s *RedisPresenceStore) GetUserSession(ctx context.Context, userID string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s", userID)
	
	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("user not online: %s", userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	var session SessionData
	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	
	return &session, nil
}

// RemoveUserSession removes user session from Redis
func (s *RedisPresenceStore) RemoveUserSession(ctx context.Context, userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	return s.client.Del(ctx, key).Err()
}

// IsUserOnline checks if user is online
func (s *RedisPresenceStore) IsUserOnline(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("session:%s", userID)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// RefreshTTL extends the session TTL (called on heartbeat)
func (s *RedisPresenceStore) RefreshTTL(ctx context.Context, userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	return s.client.Expire(ctx, key, s.ttl).Err()
}
