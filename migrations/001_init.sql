-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
    id VARCHAR(36) PRIMARY KEY,
    name TEXT NOT NULL,
    date timestamptz NOT NULL,
    total_seats INT NOT NULL,
    available INT NOT NULL,
    booking_ttl INTERVAL NOT NULL,
    requires_payment BOOLEAN NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    email TEXT NOT NULL,
    telegram TEXT,
    role TEXT NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE TABLE bookings (
    id VARCHAR(36) PRIMARY KEY,
    event_id VARCHAR(36) REFERENCES events(id),
    user_id VARCHAR(36) REFERENCES users(id),
    status TEXT NOT NULL,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    confirmed_at timestamptz
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS events;
-- +goose StatementEnd