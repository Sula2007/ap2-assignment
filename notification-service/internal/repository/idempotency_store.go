package repository

import (
	"database/sql"
	"fmt"
)

type IdempotencyStore interface {
	HasProcessed(eventID string) (bool, error)
	MarkProcessed(eventID string) error
}

type postgresIdempotencyStore struct {
	db *sql.DB
}

func NewPostgresIdempotencyStore(db *sql.DB) IdempotencyStore {
	return &postgresIdempotencyStore{db: db}
}

func (s *postgresIdempotencyStore) HasProcessed(eventID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM processed_events WHERE event_id = $1)`, eventID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check processed: %w", err)
	}
	return exists, nil
}

func (s *postgresIdempotencyStore) MarkProcessed(eventID string) error {
	_, err := s.db.Exec(
		`INSERT INTO processed_events (event_id, processed_at) VALUES ($1, NOW()) ON CONFLICT DO NOTHING`,
		eventID,
	)
	return err
}
