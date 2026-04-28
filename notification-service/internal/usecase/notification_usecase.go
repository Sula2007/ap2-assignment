package usecase

import (
	"fmt"
	"log"

	"notification-service/internal/domain"
	"notification-service/internal/repository"
)

type NotificationUsecase interface {
	HandlePaymentEvent(event domain.PaymentEvent) error
}

type notificationUsecase struct {
	store repository.IdempotencyStore
}

func NewNotificationUsecase(store repository.IdempotencyStore) NotificationUsecase {
	return &notificationUsecase{store: store}
}

func (u *notificationUsecase) HandlePaymentEvent(event domain.PaymentEvent) error {
	already, err := u.store.HasProcessed(event.EventID)
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}
	if already {
		log.Printf("[Notification] Duplicate event %s – skipping.", event.EventID)
		return nil
	}

	amountDollars := float64(event.Amount) / 100.0
	log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f",
		event.CustomerEmail, event.OrderID, amountDollars)

	if err := u.store.MarkProcessed(event.EventID); err != nil {
		return fmt.Errorf("mark processed failed: %w", err)
	}

	return nil
}
