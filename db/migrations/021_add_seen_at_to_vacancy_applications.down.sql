DROP INDEX IF EXISTS idx_vacancy_applications_unseen;
ALTER TABLE vacancy_applications DROP COLUMN IF EXISTS seen_at;
