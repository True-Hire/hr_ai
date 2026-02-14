DROP INDEX IF EXISTS idx_vacancies_country_id;
ALTER TABLE vacancies DROP COLUMN IF EXISTS country_id;
DROP TABLE IF EXISTS country_texts CASCADE;
DROP TABLE IF EXISTS countries CASCADE;
