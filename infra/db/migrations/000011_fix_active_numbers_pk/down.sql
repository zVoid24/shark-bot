-- Revert active_numbers primary key
ALTER TABLE active_numbers DROP CONSTRAINT IF EXISTS active_numbers_pkey;
ALTER TABLE active_numbers ADD PRIMARY KEY (number);
