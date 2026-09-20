-- +goose Up
CREATE TABLE mail_outbox (
    id TEXT PRIMARY KEY,
    recipient TEXT NOT NULL,
    reply_to TEXT NOT NULL DEFAULT '',
    subject TEXT NOT NULL,
    html TEXT NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX mail_outbox_pending ON mail_outbox(available_at) WHERE delivered_at IS NULL;

-- +goose Down
DROP TABLE mail_outbox;
