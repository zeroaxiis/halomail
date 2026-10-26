-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS deletion_scheduled_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS scheduled_delete_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS users_scheduled_delete_at_idx ON users(scheduled_delete_at) WHERE scheduled_delete_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS users_scheduled_delete_at_idx;
ALTER TABLE users DROP COLUMN IF EXISTS scheduled_delete_at;
ALTER TABLE users DROP COLUMN IF EXISTS deletion_scheduled_at;
