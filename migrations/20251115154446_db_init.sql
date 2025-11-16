-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS teams (
    name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    team_name TEXT NOT NULL REFERENCES teams(name),
    is_active BOOLEAN NOT NULL
);

CREATE INDEX idx_users_team_name ON users(team_name);

CREATE TABLE IF NOT EXISTS pull_requests (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES users(id),
    status TEXT NOT NULL CHECK (STATUS IN ('OPEN', 'MERGED')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    merged_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS reviewers (
    id SERIAL PRIMARY KEY,
    pull_request_id TEXT NOT NULL REFERENCES pull_requests(id),
    reviewer_id TEXT NOT NULL REFERENCES users(id)
);

CREATE UNIQUE INDEX reviewers_unique ON reviewers(pull_request_id, reviewer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

-- +goose StatementEnd
