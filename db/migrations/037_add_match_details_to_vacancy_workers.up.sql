-- Add match details to vacancy_workers
ALTER TABLE vacancy_workers ADD COLUMN IF NOT EXISTS match_percentage INT NOT NULL DEFAULT 0;
ALTER TABLE vacancy_workers ADD COLUMN IF NOT EXISTS match_score NUMERIC(6,3) NOT NULL DEFAULT 0;
ALTER TABLE vacancy_workers ADD COLUMN IF NOT EXISTS rank INT NOT NULL DEFAULT 0;

-- Ensure uniqueness to avoid duplicate matches for the same vacancy/user
CREATE UNIQUE INDEX IF NOT EXISTS idx_vacancy_workers_vacancy_user ON vacancy_workers(vacancy_id, user_id);
