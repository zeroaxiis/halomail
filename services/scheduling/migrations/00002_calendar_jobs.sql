-- +goose Up
ALTER TABLE bookings ADD COLUMN calendar_pending BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE bookings ADD COLUMN calendar_retry_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE bookings ADD COLUMN revision INT NOT NULL DEFAULT 1;
CREATE INDEX bookings_calendar_pending ON bookings(calendar_retry_at) WHERE calendar_pending;
CREATE TABLE calendar_oauth_states (
    state_hash TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL,
    verifier TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE TABLE google_credentials (
    owner_id TEXT PRIMARY KEY,
    encrypted_token TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE google_credentials;
DROP TABLE calendar_oauth_states;
ALTER TABLE bookings DROP COLUMN revision;
ALTER TABLE bookings DROP COLUMN calendar_retry_at;
ALTER TABLE bookings DROP COLUMN calendar_pending;
