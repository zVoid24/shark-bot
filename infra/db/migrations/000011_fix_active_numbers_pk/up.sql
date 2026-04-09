-- Migration to fix active_numbers primary key
-- Drop old PK on 'number' and add composite PK on '(number, platform)'

ALTER TABLE active_numbers DROP CONSTRAINT IF EXISTS active_numbers_pkey;
ALTER TABLE active_numbers ADD PRIMARY KEY (number, platform);
