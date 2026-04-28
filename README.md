# AP2 Assignment 3 – Event-Driven Architecture with Message Queues

## Architecture Overview

```
[Order Service] ──HTTP──> [Payment Service] ──RabbitMQ──> [Notification Service]
      │                         │                                  │
  order-db                payment-db                        notification-db
```

**Flow:**
1. Client sends `POST /api/v1/orders` to **Order Service**
2. Order Service calls **Payment Service** via HTTP
3. After successful DB commit, Payment Service publishes a `PaymentEvent` to RabbitMQ exchange `payments`
4. **Notification Service** consumes the message and logs:
   ```
   [Notification] Sent email to user@example.com for Order #123. Amount: $99.99
   ```

---

## How to Run

```bash
docker-compose up --build
```

All services, databases, and RabbitMQ start automatically.

- **Order Service:** http://localhost:8082
- **Payment Service:** http://localhost:8081
- **RabbitMQ Management UI:** http://localhost:15672 (guest / guest)

### Example Request

```bash
# 1. Create a payment (triggers event)
curl -X POST http://localhost:8081/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{"order_id":"order-001","amount":9999,"customer_email":"user@example.com"}'
```

Expected notification-service log:
```
[Notification] Sent email to user@example.com for Order #order-001. Amount: $99.99
```

---

## Idempotency Strategy

**Problem:** RabbitMQ guarantees *at-least-once* delivery. The same message may be redelivered if the consumer crashes before ACKing.

**Solution:** Every `PaymentEvent` carries a unique `event_id` (format: `evt_<UnixNano>`).

Before processing, the Notification Service queries:
```sql
SELECT EXISTS(SELECT 1 FROM processed_events WHERE event_id = $1)
```
- If **found** → skip silently (duplicate), send ACK.
- If **not found** → process, then insert into `processed_events`, then send ACK.

The `processed_events` table in `notification-db` serves as a persistent idempotency store, so it survives service restarts.

---

## ACK Logic

Manual acknowledgement is used (`auto-ack: false`).

| Scenario | Action |
|---|---|
| Message processed successfully | `msg.Ack(false)` |
| Unmarshal error (poison message) | `msg.Nack(false, false)` → routed to DLQ |
| Processing error, retries < 3 | `msg.Nack(false, true)` → requeued |
| Processing error, retries ≥ 3 | `msg.Nack(false, false)` → routed to DLQ |

QoS is set to `prefetch=1` so the consumer processes one message at a time, preventing message loss on crash.

---

## Reliability & Delivery Guarantees

| Feature | Implementation |
|---|---|
| **Durable queue** | `QueueDeclare(..., durable=true, ...)` – survives broker restart |
| **Persistent messages** | `DeliveryMode: amqp.Persistent` in publisher |
| **Manual ACK** | `auto-ack=false`; ACK sent only after successful processing |
| **At-least-once delivery** | Unacked messages are redelivered on consumer reconnect |
| **Prefetch = 1** | One in-flight message per consumer |

---

## Dead Letter Queue (Bonus)

**DLX Exchange:** `payments.dlx`  
**DLQ:** `payment.completed.dlq`

Messages are routed to the DLQ when:
- The message cannot be deserialized (poison message)
- Processing fails after **3 retry attempts**

The queue is configured with `x-dead-letter-exchange` and `x-max-delivery-count=3`.  
You can inspect dead-lettered messages in the RabbitMQ Management UI under the `payment.completed.dlq` queue.

**To simulate DLQ routing:** temporarily make the Notification Service return an error for a specific `order_id`.

---

## Graceful Shutdown

Both HTTP services (Order, Payment) and the Notification consumer use `os/signal` to catch `SIGINT`/`SIGTERM`:

- **HTTP services:** `http.Server.Shutdown(ctx)` with a 10-second timeout
- **Notification consumer:** context cancellation stops the `consumer.Start()` loop; in-flight message is NACKed and requeued before exit

---

## Project Structure

```
ap2/
├── docker-compose.yml
├── order-service/
│   ├── Dockerfile
│   ├── cmd/order-service/main.go       ← graceful shutdown, env vars
│   └── internal/...
├── payment-service/
│   ├── Dockerfile
│   ├── cmd/payment-service/main.go     ← graceful shutdown, publisher wiring
│   └── internal/
│       ├── domain/event.go             ← PaymentEvent struct
│       ├── messaging/
│       │   ├── publisher.go            ← Publisher interface (port)
│       │   └── rabbitmq_publisher.go   ← RabbitMQ implementation
│       └── usecase/payment_usecase.go  ← publishes after DB commit
└── notification-service/
    ├── Dockerfile
    ├── migrations/001_create_processed_events.sql
    ├── cmd/notification-service/main.go
    └── internal/
        ├── domain/event.go
        ├── repository/idempotency_store.go  ← PostgreSQL idempotency store
        ├── usecase/notification_usecase.go  ← duplicate check + email log
        └── messaging/consumer.go            ← RabbitMQ consumer, manual ACKs, DLQ
```
