DROP INDEX IF EXISTS users_telegram_unique;
ALTER TABLE users DROP COLUMN IF EXISTS telegram_id;
