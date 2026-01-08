# Service Responsibilities

## WebSocket Gateway

### Owns
- WebSocket connection management
- Connection registry (via Redis)
- Message routing to clients
- Connection health checks

### Does NOT Own
- Message persistence
- Business logic
- Authorization (validates JWT only)

### Dependencies
- Redis (session store)
- Chat Service (via HTTP/gRPC)

## Chat Service

### Owns
- Message validation
- Authorization logic
- Database persistence
- Kafka publishing

### Does NOT Own
- WebSocket connections
- Message delivery
- Session management

### Dependencies
- Database (MongoDB/NoSQL)
- Kafka (producer)
- Gateway (via service discovery)

## Delivery Worker

### Owns
- Kafka consumption
- Delivery status updates
- Idempotent message delivery

### Does NOT Own
- Message creation
- WebSocket connections

### Dependencies
- Kafka (consumer)
- Gateway (delivery API)
- Database (status updates)

## Clear Boundaries

```
[Gateway]     [Chat Service]     [Delivery Worker]
   |               |                     |
Connections    Persistence          Delivery
   |               |                     |
Redis          Database              Kafka
```
