# System Architecture Overview

## Core Components

### 1. WebSocket Gateway
**Responsibility**: Manage client connections
- Handle WebSocket lifecycle
- Route incoming messages
- Maintain connection registry
- Stateless (connections stored in Redis)

### 2. Chat Service
**Responsibility**: Business logic and persistence
- Validate message authorization
- Persist messages to database
- Publish events to Kafka
- Stateless processing

### 3. Delivery Worker
**Responsibility**: Message delivery
- Consume from Kafka
- Deliver to online users via Gateway
- Update delivery status
- Idempotent processing

### 4. Redis Cache
**Responsibility**: Session and routing data
- Online user presence
- Connection-to-gateway mapping
- Hot message cache

### 5. Database (NoSQL)
**Responsibility**: Source of truth
- Message persistence
- User data
- Delivery status

### 6. Kafka
**Responsibility**: Event streaming
- Message events
- Delivery events
- Status updates

## Data Flow

```
Client A                  Client B
  |                          |
  v                          v
[WS Gateway] <---------> [WS Gateway]
  |                          ^
  v                          |
[Chat Service] ---------> [Kafka]
  |                          |
  v                          v
[Database]             [Delivery Worker]
```

## Design Principles

1. **Separation of Concerns**: Each service has single responsibility
2. **Stateless Services**: All state in Redis/DB
3. **Event-Driven**: Async via Kafka
4. **Database as Truth**: All writes persisted first
5. **Idempotency**: Workers handle duplicate events

## Failure Handling

- Gateway crash: Reconnect, fetch missed messages
- Service crash: Stateless restart
- Kafka lag: Eventual delivery
- Database failure: Service unavailable (no data loss)
