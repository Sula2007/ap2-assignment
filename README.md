# AP2 Assignment 2 — gRPC Migration
**Student:** Moldash Sultan | **Group:** SE-2405

## Repository Links
- **Proto Repository:** https://github.com/Sula2007/ap2-protos
- **Generated Code Repository:** https://github.com/Sula2007/ap2-generated

## Architecture

```
[Client/User]
     |
     | HTTP REST
     v
[Order Service :8082] -------gRPC PaymentService--------> [Payment Service :50051]
     |                                                           |
     |                                                     saves to DB
     v                                                           v
[Order DB :5432]                                         [Payment DB :5433]

[grpcurl / test client] ---gRPC Streaming--> [Order Service :50052]
                                              (polls DB, pushes status changes)
```

### What changed from Assignment 1
- Order Service calls Payment Service via **gRPC** (was HTTP REST)
- Payment Service runs a **gRPC server** on :50051 alongside HTTP :8081
- Order Service runs a **gRPC streaming server** on :50052 for real-time order status
- All addresses/ports moved to **.env files** (no hardcoding)
- Payment Service has a **gRPC Logging Interceptor** (+10% bonus)

## How to Run

```bash
docker-compose up --build
```

| Service | HTTP | gRPC |
|---------|------|------|
| Order Service | :8082 | :50052 (streaming) |
| Payment Service | :8081 | :50051 |

## API Examples

### Create Order
```bash
curl -X POST http://localhost:8082/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"c1","item_name":"Laptop","amount":50000}'
```

### Get Order
```bash
curl http://localhost:8082/api/v1/orders/{id}
```

### Cancel Order
```bash
curl -X PATCH http://localhost:8082/api/v1/orders/{id}/cancel
```

### Subscribe to real-time Order Status Stream
```bash
grpcurl -plaintext \
  -d '{"order_id":"YOUR_ORDER_ID"}' \
  localhost:50052 \
  order.OrderService/SubscribeToOrderUpdates
```

## Contract-First Flow

1. `.proto` files in [ap2-protos](https://github.com/Sula2007/ap2-protos)
2. GitHub Actions auto-generates `.pb.go` and pushes to [ap2-generated](https://github.com/Sula2007/ap2-generated)
3. Services import: `go get github.com/Sula2007/ap2-generated@v1.0.0`
