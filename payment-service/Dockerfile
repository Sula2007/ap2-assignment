FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o payment-service ./cmd/payment-service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/payment-service .
COPY .env .
EXPOSE 8081 50051
CMD ["./payment-service"]
