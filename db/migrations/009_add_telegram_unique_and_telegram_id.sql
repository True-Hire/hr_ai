ALTER TABLE users ADD COLUMN IF NOT EXISTS telegram_id TEXT UNIQUE;
ALTER TABLE users ADD CONSTRAINT users_telegram_unique UNIQUE (telegram);
