DROP INDEX IF EXISTS idx_csp_sub_category_id;
DROP INDEX IF EXISTS idx_csp_main_category_id;

ALTER TABLE candidate_search_profiles DROP COLUMN IF EXISTS sub_category_id;
ALTER TABLE candidate_search_profiles DROP COLUMN IF EXISTS main_category_id;
