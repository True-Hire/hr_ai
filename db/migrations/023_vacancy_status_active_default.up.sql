ALTER TABLE vacancies ALTER COLUMN status SET DEFAULT 'active';
UPDATE vacancies SET status = 'active' WHERE status = 'draft';
