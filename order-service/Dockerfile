FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN go build -o order-service ./cmd/order-service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/order-service .
COPY .env .
EXPOSE 8082 50052
CMD ["./order-service"]
