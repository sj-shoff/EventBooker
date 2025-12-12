package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"event-booker/internal/domain"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type BookingRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewBookingRepository(db *dbpg.DB, retries retry.Strategy) *BookingRepository {
	return &BookingRepository{db: db, retries: retries}
}

func (r *BookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	query := `
	INSERT INTO bookings (id, event_id, user_id, status, created_at, expires_at, confirmed_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecWithRetry(ctx, r.retries, query,
		booking.ID, booking.EventID, booking.UserID, booking.Status, booking.CreatedAt, booking.ExpiresAt, booking.ConfirmedAt)
	return err
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	query := `
	SELECT id, event_id, user_id, status, created_at, expires_at, confirmed_at
	FROM bookings WHERE id = $1
	`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return nil, err
	}
	var booking domain.Booking
	err = row.Scan(&booking.ID, &booking.EventID, &booking.UserID, &booking.Status, &booking.CreatedAt, &booking.ExpiresAt, &booking.ConfirmedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("booking not found")
	}
	return &booking, err
}

func (r *BookingRepository) Update(ctx context.Context, booking *domain.Booking) error {
	query := `
	UPDATE bookings SET status = $1, confirmed_at = $2 WHERE id = $3
	`
	_, err := r.db.ExecWithRetry(ctx, r.retries, query, booking.Status, booking.ConfirmedAt, booking.ID)
	return err
}

func (r *BookingRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM bookings WHERE id = $1`
	_, err := r.db.ExecWithRetry(ctx, r.retries, query, id)
	return err
}

func (r *BookingRepository) GetExpired(ctx context.Context, now time.Time) ([]*domain.Booking, error) {
	query := `
	SELECT id, event_id, user_id, status, created_at, expires_at, confirmed_at
	FROM bookings WHERE status = 'pending' AND expires_at < $1
	`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bookings []*domain.Booking
	for rows.Next() {
		var b domain.Booking
		err := rows.Scan(&b.ID, &b.EventID, &b.UserID, &b.Status, &b.CreatedAt, &b.ExpiresAt, &b.ConfirmedAt)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, &b)
	}
	return bookings, nil
}

func (r *BookingRepository) GetByEventID(ctx context.Context, eventID string) ([]*domain.Booking, error) {
	query := `
	SELECT id, event_id, user_id, status, created_at, expires_at, confirmed_at
	FROM bookings WHERE event_id = $1
	`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bookings []*domain.Booking
	for rows.Next() {
		var b domain.Booking
		err := rows.Scan(&b.ID, &b.EventID, &b.UserID, &b.Status, &b.CreatedAt, &b.ExpiresAt, &b.ConfirmedAt)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, &b)
	}
	return bookings, nil
}
