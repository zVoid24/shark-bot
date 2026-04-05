CREATE TABLE IF NOT EXISTS processed_otp_events (
    phone_number TEXT NOT NULL,
    otp_code TEXT NOT NULL,
    service_name TEXT,
    posted BOOLEAN DEFAULT TRUE,
    first_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (phone_number, otp_code)
);

CREATE INDEX IF NOT EXISTS idx_processed_otp_events_phone ON processed_otp_events(phone_number);
