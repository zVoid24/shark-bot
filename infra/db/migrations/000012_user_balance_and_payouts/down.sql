DROP TABLE IF EXISTS earnings_log;
ALTER TABLE users DROP COLUMN IF EXISTS balance;
ALTER TABLE active_numbers DROP COLUMN IF EXISTS payout_done;
