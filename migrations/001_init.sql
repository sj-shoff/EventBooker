-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
    id VARCHAR(36) PRIMARY KEY,
    name TEXT NOT NULL,
    date timestamptz NOT NULL,
    total_seats INT NOT NULL,
    available INT NOT NULL,
    booking_ttl INTERVAL NOT NULL,
    requires_payment BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT events_status_check CHECK (status IN ('active', 'cancelled', 'completed'))
);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_date ON events(date);

CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    telegram TEXT,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_role_check CHECK (role IN ('user', 'admin'))
);
CREATE INDEX idx_users_email ON users(email);

CREATE TABLE bookings (
    id VARCHAR(36) PRIMARY KEY,
    event_id VARCHAR(36) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    confirmed_at timestamptz,
    CONSTRAINT bookings_status_check CHECK (status IN ('pending', 'confirmed', 'cancelled')),
    CONSTRAINT bookings_event_user_unique UNIQUE (event_id, user_id)
);
CREATE INDEX idx_bookings_event_id ON bookings(event_id);
CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_expires_at ON bookings(expires_at) WHERE status = 'pending';
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS events;
-- +goose StatementEnd