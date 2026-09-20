-- +goose Up
CREATE UNIQUE INDEX api_keys_one_active_feature ON api_keys (user_id, (scopes[1]))
WHERE NOT revoked AND cardinality(scopes) = 1 AND scopes[1] IN ('forms:submit', 'meetings:book');

CREATE TABLE feature_usage (
    owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feature TEXT NOT NULL CHECK (feature IN ('forms', 'meetings')),
    period DATE NOT NULL,
    used INT NOT NULL DEFAULT 0 CHECK (used >= 0),
    PRIMARY KEY (owner_id, feature, period)
);

-- +goose Down
DROP TABLE feature_usage;
DROP INDEX api_keys_one_active_feature;
