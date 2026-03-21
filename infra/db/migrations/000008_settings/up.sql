CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);

-- Default settings seed
INSERT INTO settings (key, value)
VALUES ('group_link', 'https://t.me/tgwscreatebdotp')
ON CONFLICT (key) DO NOTHING;
