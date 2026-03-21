CREATE TABLE IF NOT EXISTS seen_numbers (
    user_id TEXT NOT NULL,
    number  TEXT NOT NULL,
    country TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_seen_user ON seen_numbers(user_id, country);
