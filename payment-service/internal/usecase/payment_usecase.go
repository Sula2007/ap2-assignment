package usecase

import (
	"errors"
	"fmt"
	"time"

	"payment-service/internal/domain"
	"payment-service/internal/repository"
)

type PaymentUsecase interface {
	AuthorizePayment(orderID string, amount int64) (*domain.Payment, error)
	GetPaymentByOrderID(orderID string) (*domain.Payment, error)
	ListPayments(min, max int64) ([]*domain.Payment, error)
}

type paymentUsecase struct {
	repo repository.PaymentRepository
}

func NewPaymentUsecase(repo repository.PaymentRepository) PaymentUsecase {
	return &paymentUsecase{repo: repo}
}

func (u *paymentUsecase) AuthorizePayment(orderID string, amount int64) (*domain.Payment, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	status := "Authorized"
	if amount > 100000 {
		status = "Declined"
	}

	payment := &domain.Payment{
		ID:            fmt.Sprintf("%d", time.Now().UnixNano()),
		OrderID:       orderID,
		TransactionID: fmt.Sprintf("txn_%d", time.Now().UnixNano()),
		Amount:        amount,
		Status:        status,
		CreatedAt:     time.Now(),
	}

	if err := u.repo.Create(payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (u *paymentUsecase) GetPaymentByOrderID(orderID string) (*domain.Payment, error) {
	payment, err := u.repo.FindByOrderID(orderID)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, fmt.Errorf("payment not found for order: %s", orderID)
	}
	return payment, nil
}

func (u *paymentUsecase) ListPayments(min, max int64) ([]*domain.Payment, error) {
	if min > 0 && max > 0 && min > max {
		return nil, errors.New("min_amount cannot be greater than max_amount")
	}
	return u.repo.FindByAmountRange(min, max)
}