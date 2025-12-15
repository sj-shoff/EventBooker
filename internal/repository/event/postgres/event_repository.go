package event_postgres

import (
	"context"
	"database/sql"
	"time"

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
INSERT INTO events (id, name, date, total_seats, available, booking_ttl, requires_payment, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`
	_, err := r.db.ExecWithRetry(ctx, r.retries, query,
		event.ID, event.Name, event.Date, event.TotalSeats, event.Available,
		event.BookingTTL, event.RequiresPayment, event.Status, event.CreatedAt, event.UpdatedAt)
	return err
}

func (r *EventRepository) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	query := `
SELECT id, name, date, total_seats, available, booking_ttl, requires_payment, status, created_at, updated_at
FROM events WHERE id = $1
`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return nil, err
	}

	var event domain.Event
	var bookingTTLStr string
	var status sql.NullString

	err = row.Scan(
		&event.ID,
		&event.Name,
		&event.Date,
		&event.TotalSeats,
		&event.Available,
		&bookingTTLStr,
		&event.RequiresPayment,
		&status,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	duration, err := time.ParseDuration(bookingTTLStr)
	if err != nil {
		event.BookingTTL = 30 * time.Minute
	} else {
		event.BookingTTL = duration
	}

	if status.Valid {
		event.Status = domain.EventStatus(status.String)
	} else {
		event.Status = domain.EventActive
	}

	return &event, nil
}

func (r *EventRepository) GetForUpdate(ctx context.Context, tx *sql.Tx, id string) (*domain.Event, error) {
	query := `
SELECT id, name, date, total_seats, available, booking_ttl, requires_payment, status, created_at, updated_at
FROM events WHERE id = $1 FOR UPDATE
`
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, id)
	} else {
		rowResult, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
		if err != nil {
			return nil, err
		}
		row = rowResult
	}

	var event domain.Event
	var bookingTTLStr string
	var status sql.NullString

	err := row.Scan(
		&event.ID,
		&event.Name,
		&event.Date,
		&event.TotalSeats,
		&event.Available,
		&bookingTTLStr,
		&event.RequiresPayment,
		&status,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	duration, err := time.ParseDuration(bookingTTLStr)
	if err != nil {
		event.BookingTTL = 30 * time.Minute
	} else {
		event.BookingTTL = duration
	}

	if status.Valid {
		event.Status = domain.EventStatus(status.String)
	} else {
		event.Status = domain.EventActive
	}

	return &event, nil
}

func (r *EventRepository) GetAll(ctx context.Context) ([]*domain.Event, error) {
	query := `
SELECT id, name, date, total_seats, available, booking_ttl, requires_payment, status, created_at, updated_at
FROM events
ORDER BY date ASC, created_at DESC
`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.Event

	for rows.Next() {
		var event domain.Event
		var bookingTTLStr string
		var status sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.Date,
			&event.TotalSeats,
			&event.Available,
			&bookingTTLStr,
			&event.RequiresPayment,
			&status,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		duration, err := time.ParseDuration(bookingTTLStr)
		if err != nil {
			event.BookingTTL = 30 * time.Minute
		} else {
			event.BookingTTL = duration
		}

		if status.Valid {
			event.Status = domain.EventStatus(status.String)
		} else {
			event.Status = domain.EventActive
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepository) Update(ctx context.Context, tx *sql.Tx, event *domain.Event) error {
	query := `
UPDATE events 
SET name = $1, date = $2, total_seats = $3, available = $4, 
    booking_ttl = $5, requires_payment = $6, status = $7, updated_at = $8
WHERE id = $9
`
	if tx != nil {
		_, err := tx.ExecContext(ctx, query,
			event.Name, event.Date, event.TotalSeats, event.Available,
			event.BookingTTL, event.RequiresPayment, event.Status, event.UpdatedAt, event.ID)
		return err
	}
	_, err := r.db.ExecWithRetry(ctx, r.retries, query,
		event.Name, event.Date, event.TotalSeats, event.Available,
		event.BookingTTL, event.RequiresPayment, event.Status, event.UpdatedAt, event.ID)
	return err
}

func (r *EventRepository) Delete(ctx context.Context, eventID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteBookingsQuery := `DELETE FROM bookings WHERE event_id = $1`
	_, err = tx.ExecContext(ctx, deleteBookingsQuery, eventID)
	if err != nil {
		return err
	}

	deleteEventQuery := `DELETE FROM events WHERE id = $1`
	_, err = tx.ExecContext(ctx, deleteEventQuery, eventID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *EventRepository) DecrementAvailableSeats(ctx context.Context, tx *sql.Tx, id string) error {
	query := `UPDATE events SET available = available - 1, updated_at = NOW() WHERE id = $1 AND available > 0`
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, id)
		return err
	}
	_, err := r.db.ExecWithRetry(ctx, r.retries, query, id)
	return err
}

func (r *EventRepository) IncrementAvailableSeats(ctx context.Context, tx *sql.Tx, id string) error {
	query := `UPDATE events SET available = available + 1, updated_at = NOW() WHERE id = $1`
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, id)
		return err
	}
	_, err := r.db.ExecWithRetry(ctx, r.retries, query, id)
	return err
}

func (r *EventRepository) GetActiveEvents(ctx context.Context) ([]*domain.Event, error) {
	query := `
SELECT id, name, date, total_seats, available, booking_ttl, requires_payment, status, created_at, updated_at
FROM events 
WHERE status = 'active' AND date >= NOW()
ORDER BY date ASC
`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.Event

	for rows.Next() {
		var event domain.Event
		var bookingTTLStr string
		var status sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.Date,
			&event.TotalSeats,
			&event.Available,
			&bookingTTLStr,
			&event.RequiresPayment,
			&status,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		duration, err := time.ParseDuration(bookingTTLStr)
		if err != nil {
			event.BookingTTL = 30 * time.Minute
		} else {
			event.BookingTTL = duration
		}

		if status.Valid {
			event.Status = domain.EventStatus(status.String)
		} else {
			event.Status = domain.EventActive
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepository) GetEventsWithPagination(ctx context.Context, limit, offset int) ([]*domain.Event, error) {
	query := `
SELECT id, name, date, total_seats, available, booking_ttl, requires_payment, status, created_at, updated_at
FROM events
ORDER BY date ASC, created_at DESC
LIMIT $1 OFFSET $2
`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.Event

	for rows.Next() {
		var event domain.Event
		var bookingTTLStr string
		var status sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.Date,
			&event.TotalSeats,
			&event.Available,
			&bookingTTLStr,
			&event.RequiresPayment,
			&status,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		duration, err := time.ParseDuration(bookingTTLStr)
		if err != nil {
			event.BookingTTL = 30 * time.Minute
		} else {
			event.BookingTTL = duration
		}

		if status.Valid {
			event.Status = domain.EventStatus(status.String)
		} else {
			event.Status = domain.EventActive
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepository) CountEvents(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM events`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query)
	if err != nil {
		return 0, err
	}

	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *EventRepository) CountActiveEvents(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM events WHERE status = 'active' AND date >= NOW()`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query)
	if err != nil {
		return 0, err
	}

	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *EventRepository) GetEventsByStatus(ctx context.Context, status domain.EventStatus) ([]*domain.Event, error) {
	query := `
SELECT id, name, date, total_seats, available, booking_ttl, requires_payment, status, created_at, updated_at
FROM events 
WHERE status = $1
ORDER BY date ASC
`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.Event

	for rows.Next() {
		var event domain.Event
		var bookingTTLStr string
		var dbStatus sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.Date,
			&event.TotalSeats,
			&event.Available,
			&bookingTTLStr,
			&event.RequiresPayment,
			&dbStatus,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		duration, err := time.ParseDuration(bookingTTLStr)
		if err != nil {
			event.BookingTTL = 30 * time.Minute
		} else {
			event.BookingTTL = duration
		}

		if dbStatus.Valid {
			event.Status = domain.EventStatus(dbStatus.String)
		} else {
			event.Status = domain.EventActive
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
