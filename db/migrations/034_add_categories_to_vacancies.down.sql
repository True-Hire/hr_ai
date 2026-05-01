DROP INDEX IF EXISTS idx_vacancies_sub_category_id;
DROP INDEX IF EXISTS idx_vacancies_main_category_id;
ALTER TABLE vacancies DROP COLUMN IF EXISTS sub_category_id;
ALTER TABLE vacancies DROP COLUMN IF EXISTS main_category_id;
