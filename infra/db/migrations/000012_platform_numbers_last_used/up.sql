DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='platform_numbers' AND column_name='last_used_at') THEN
        ALTER TABLE platform_numbers ADD COLUMN last_used_at TIMESTAMP DEFAULT NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_pn_last_used ON platform_numbers(last_used_at);
