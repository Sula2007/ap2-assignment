package repository

import (
	"database/sql"
	"fmt"

	"payment-service/internal/domain"
)

type PaymentRepository interface {
	Create(payment *domain.Payment) error
	FindByOrderID(orderID string) (*domain.Payment, error)
	FindByAmountRange(min, max int64) ([]*domain.Payment, error)
}

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *domain.Payment) error {
	query := `INSERT INTO payments (id, order_id, transaction_id, amount, status, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, payment.ID, payment.OrderID, payment.TransactionID,
		payment.Amount, payment.Status, payment.CreatedAt)
	return err
}

func (r *paymentRepository) FindByOrderID(orderID string) (*domain.Payment, error) {
	query := `SELECT id, order_id, transaction_id, amount, status, created_at
	          FROM payments WHERE order_id = $1`

	var payment domain.Payment
	err := r.db.QueryRow(query, orderID).Scan(
		&payment.ID, &payment.OrderID, &payment.TransactionID,
		&payment.Amount, &payment.Status, &payment.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) FindByAmountRange(min, max int64) ([]*domain.Payment, error) {
	query := `SELECT id, order_id, transaction_id, amount, status, created_at FROM payments WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if min > 0 {
		query += fmt.Sprintf(" AND amount >= $%d", idx)
		args = append(args, min)
		idx++
	}
	if max > 0 {
		query += fmt.Sprintf(" AND amount <= $%d", idx)
		args = append(args, max)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, &p)
	}
	return payments, nil
}
