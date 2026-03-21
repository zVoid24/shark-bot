CREATE TABLE IF NOT EXISTS user_otp_stats (
    user_id TEXT NOT NULL,
    country TEXT NOT NULL,
    count   INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, country)
);
