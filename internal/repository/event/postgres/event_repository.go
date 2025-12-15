package event_postgres

import (
	"context"
	"database/sql"

	"event-booker/internal/domain"
	"event-booker/internal/repository"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type EventRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewEventRepository(db *dbpg.DB, retries retry.Strategy) *EventRepository {
	return &EventRepository{db: db, retries: retries}
}

func (r *EventRepository) Create(ctx context.Context, event *domain.Event) error {
	query := `
INSERT INTO events (id, name, date, total_seats, available, booking_ttl, requires_payment, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`
	_, err := r.db.ExecWithRetry(ctx, r.retries, query,
		event.ID, event.Name, event.Date, event.TotalSeats, event.Available, event.BookingTTL, event.RequiresPayment, event.CreatedAt, event.UpdatedAt)
	return err
}

func (r *EventRepository) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	query := `
SELECT id, name, date, total_seats, available, booking_ttl, requires_payment, created_at, updated_at
FROM events WHERE id = $1
`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return nil, err
	}
	var event domain.Event
	err = row.Scan(&event.ID, &event.Name, &event.Date, &event.TotalSeats, &event.Available, &event.BookingTTL, &event.RequiresPayment, &event.CreatedAt, &event.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) GetAll(ctx context.Context) ([]*domain.Event, error) {
	query := `
SELECT id, name, date, total_seats, available, booking_ttl, requires_payment, created_at, updated_at
FROM events
`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []*domain.Event
	for rows.Next() {
		var e domain.Event
		err := rows.Scan(&e.ID, &e.Name, &e.Date, &e.TotalSeats, &e.Available, &e.BookingTTL, &e.RequiresPayment, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepository) GetForUpdate(ctx context.Context, tx *sql.Tx, id string) (*domain.Event, error) {
	query := `
SELECT id, name, date, total_seats, available, booking_ttl, requires_payment, created_at, updated_at
FROM events WHERE id = $1 FOR UPDATE
`
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, id)
	} else {
		row, _ = r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	}
	var event domain.Event
	err := row.Scan(&event.ID, &event.Name, &event.Date, &event.TotalSeats, &event.Available, &event.BookingTTL, &event.RequiresPayment, &event.CreatedAt, &event.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) DecrementAvailableSeats(ctx context.Context, tx *sql.Tx, id string) error {
	query := `UPDATE events SET available = available - 1 WHERE id = $1 AND available > 0`
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, id)
		return err
	}
	_, err := r.db.ExecWithRetry(ctx, r.retries, query, id)
	return err
}

func (r *EventRepository) IncrementAvailableSeats(ctx context.Context, tx *sql.Tx, id string) error {
	query := `UPDATE events SET available = available + 1 WHERE id = $1`
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, id)
		return err
	}
	_, err := r.db.ExecWithRetry(ctx, r.retries, query, id)
	return err
}
