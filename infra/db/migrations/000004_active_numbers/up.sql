CREATE TABLE IF NOT EXISTS active_numbers (
    number     TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    timestamp  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    message_id BIGINT NOT NULL DEFAULT 0,
    platform   TEXT NOT NULL DEFAULT '',
    country    TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_an_user ON active_numbers(user_id);
