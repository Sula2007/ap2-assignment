package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"notification-service/internal/messaging"
	"notification-service/internal/repository"
	"notification-service/internal/usecase"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	dsn := getEnv("DATABASE_URL",
		"host=localhost port=5434 user=postgres password=postgres dbname=notification_db sslmode=disable")
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer db.Close()

	for i := 0; i < 10; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		log.Printf("DB not ready, retrying (%d/10)...", i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatal("DB ping failed:", err)
	}

	idempotencyStore := repository.NewPostgresIdempotencyStore(db)
	notifUsecase := usecase.NewNotificationUsecase(idempotencyStore)

	consumer, err := messaging.NewConsumer(rabbitmqURL, notifUsecase)
	if err != nil {
		log.Fatal("Failed to create consumer:", err)
	}
	defer consumer.Close()

	// ── Graceful Shutdown ────────────────────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("[Notification] Shutting down gracefully...")
		cancel()
	}()

	log.Println("Notification service started.")
	if err := consumer.Start(ctx); err != nil {
		log.Fatal("Consumer error:", err)
	}
	log.Println("[Notification] Service stopped.")
}
