-- Add category columns to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS main_category_id UUID REFERENCES main_category(id) ON DELETE SET NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS sub_category_id UUID REFERENCES sub_category(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_users_main_category_id ON users(main_category_id);
CREATE INDEX IF NOT EXISTS idx_users_sub_category_id ON users(sub_category_id);
