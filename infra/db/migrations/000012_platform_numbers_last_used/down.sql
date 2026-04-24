DROP INDEX IF EXISTS idx_pn_last_used;
-- SQLite does not support dropping columns easily, but we can leave it or ignore it.
-- For a real migration system we'd need a table recreation, but here we just append.
