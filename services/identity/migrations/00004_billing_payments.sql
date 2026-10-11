-- +goose Up
CREATE TABLE billing_payments (
    payment_id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL UNIQUE,
    org_id TEXT NOT NULL REFERENCES orgs(id),
    tier TEXT NOT NULL,
    paid_at TIMESTAMPTZ NOT NULL,
    cycle_end TIMESTAMPTZ NOT NULL
);
ALTER TABLE orgs ADD COLUMN billing_paid_at TIMESTAMPTZ;
-- +goose Down
ALTER TABLE orgs DROP COLUMN billing_paid_at;
DROP TABLE billing_payments;
