-- Add category columns to candidate_search_profiles for strict filtering
ALTER TABLE candidate_search_profiles ADD COLUMN IF NOT EXISTS main_category_id UUID REFERENCES main_category(id) ON DELETE SET NULL;
ALTER TABLE candidate_search_profiles ADD COLUMN IF NOT EXISTS sub_category_id UUID REFERENCES sub_category(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_csp_main_category_id ON candidate_search_profiles(main_category_id);
CREATE INDEX IF NOT EXISTS idx_csp_sub_category_id ON candidate_search_profiles(sub_category_id);
