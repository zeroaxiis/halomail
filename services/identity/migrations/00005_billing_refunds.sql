-- +goose Up
-- Kept independently of billing_payments so refunds can arrive before capture callbacks.
CREATE TABLE billing_refunds (
    refund_id TEXT PRIMARY KEY,
    payment_id TEXT NOT NULL,
    order_id TEXT NOT NULL,
    org_id TEXT NOT NULL REFERENCES orgs(id),
    payment_amount BIGINT NOT NULL CHECK (payment_amount > 0),
    amount BIGINT NOT NULL CHECK (amount > 0 AND amount <= payment_amount),
    currency TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX billing_refunds_payment_idx ON billing_refunds(payment_id);
CREATE INDEX billing_refunds_order_idx ON billing_refunds(order_id);

-- +goose Down
DROP TABLE billing_refunds;
