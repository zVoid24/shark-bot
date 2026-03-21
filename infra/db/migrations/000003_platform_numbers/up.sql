CREATE TABLE IF NOT EXISTS platform_numbers (
    id       SERIAL PRIMARY KEY,
    platform TEXT NOT NULL,
    country  TEXT NOT NULL,
    number   TEXT NOT NULL,
    UNIQUE (platform, country, number)
);
CREATE INDEX IF NOT EXISTS idx_pn_plat_coun ON platform_numbers(platform, country);
CREATE INDEX IF NOT EXISTS idx_pn_number    ON platform_numbers(number);
