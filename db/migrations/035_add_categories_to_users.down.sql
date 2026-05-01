DROP INDEX IF EXISTS idx_users_sub_category_id;
DROP INDEX IF EXISTS idx_users_main_category_id;
ALTER TABLE users DROP COLUMN IF EXISTS sub_category_id;
ALTER TABLE users DROP COLUMN IF EXISTS main_category_id;
