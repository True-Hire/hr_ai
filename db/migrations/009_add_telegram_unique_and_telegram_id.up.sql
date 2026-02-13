ALTER TABLE users ADD COLUMN IF NOT EXISTS telegram_id TEXT UNIQUE;
CREATE UNIQUE INDEX IF NOT EXISTS users_telegram_unique ON users(telegram);
