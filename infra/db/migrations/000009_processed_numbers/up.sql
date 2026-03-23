CREATE TABLE IF NOT EXISTS processed_numbers (
    phone_number TEXT PRIMARY KEY,
    first_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    otp_code TEXT,
    service_name TEXT,
    posted BOOLEAN DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_processed_phone ON processed_numbers(phone_number);
