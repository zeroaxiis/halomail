-- +goose Up
-- +goose StatementBegin
ALTER TABLE orgs ADD COLUMN tier TEXT NOT NULL DEFAULT 'free';
ALTER TABLE orgs ADD COLUMN rzp_customer_id TEXT;
ALTER TABLE orgs ADD COLUMN rzp_subscription_id TEXT;
ALTER TABLE orgs ADD COLUMN billing_cycle_end TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orgs DROP COLUMN tier;
ALTER TABLE orgs DROP COLUMN rzp_customer_id;
ALTER TABLE orgs DROP COLUMN rzp_subscription_id;
ALTER TABLE orgs DROP COLUMN billing_cycle_end;
-- +goose StatementEnd
