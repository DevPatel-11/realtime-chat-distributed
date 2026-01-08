# WebSocket Connection Lifecycle

## Connection Flow

```
Client                  Gateway                Redis
  |                       |                      |
  |--WebSocket Connect--->|                      |
  |                       |                      |
  |<--Challenge (401)-----|                      |
  |                       |                      |
  |--JWT Token----------->|                      |
  |                       |--Validate----------->|
  |                       |                      |
  |<--Connected (200)-----|                      |
  |                       |--Store Session------>|
  |                       |  (user_id->conn_id)  |
```

## Handshake Steps

1. Client initiates WebSocket connection
2. Gateway responds with 401, requiring JWT
3. Client sends JWT in header/message
4. Gateway validates JWT
5. On success:
   - Generate connection_id
   - Store in Redis: `session:{user_id} -> {conn_id, gateway_addr, timestamp}`
   - Send connection success
6. On failure: Close connection with error

## Presence Tracking

### Redis Schema
```
Key: session:{user_id}
Value: {
  "connection_id": "uuid",
  "gateway_addr": "gateway-1:8080",
  "connected_at": "timestamp"
}
TTL: 30 minutes (refreshed via heartbeat)
```

### Heartbeat
- Client sends ping every 30s
- Gateway responds with pong
- Gateway refreshes Redis TTL
- Miss 3 consecutive pings -> disconnect

## Reconnection Scenarios

### Clean Reconnect
1. Client detects disconnect
2. Wait exponential backoff (1s, 2s, 4s...)
3. Initiate new connection
4. Gateway removes old session, creates new

### Duplicate Connection
- If user_id already has active session:
  - Close OLD connection
  - Accept NEW connection
  - Update Redis with new connection_id

### Token Expiration
- Gateway checks JWT expiry on every message
- If expired:
  - Send "token_expired" event
  - Client must refresh token
  - Reconnect with new token

## Disconnection Cleanup

1. WebSocket close event triggered
2. Gateway:
   - Remove from Redis: `DEL session:{user_id}`
   - Remove from local connection map
   - Log disconnect event
3. No message delivery attempted to offline users

## Edge Cases

### Gateway Crash
- Redis TTL expires (30 min)
- Client detects disconnect, reconnects
- New gateway handles connection

### Network Partition
- Client thinks connected
- Gateway sees disconnect
- Client fails to receive messages
- Heartbeat timeout triggers client reconnect

### Simultaneous Connections
- User opens multiple tabs
- Each gets unique connection_id
- Redis stores most recent
- Only ONE active delivery target per user
