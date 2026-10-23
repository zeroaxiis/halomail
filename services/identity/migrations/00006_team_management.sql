-- +goose Up
-- Add role to users table with 'owner' default for existing accounts
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'owner'
    CHECK (role IN ('owner', 'admin', 'manager', 'viewer'));

CREATE TABLE org_invitations (
    id         TEXT PRIMARY KEY,
    org_id     TEXT NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    email      TEXT NOT NULL,
    role       TEXT NOT NULL CHECK (role IN ('admin', 'manager', 'viewer')),
    token      TEXT NOT NULL UNIQUE,
    invited_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'revoked', 'expired')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX org_invitations_org_idx ON org_invitations(org_id, status);
CREATE INDEX org_invitations_token_idx ON org_invitations(token);
CREATE INDEX org_invitations_email_idx ON org_invitations(lower(email));

-- +goose Down
DROP TABLE IF EXISTS org_invitations;
ALTER TABLE users DROP COLUMN IF EXISTS role;
