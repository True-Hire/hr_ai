ALTER TABLE vacancy_applications ADD COLUMN seen_at TIMESTAMP;

CREATE INDEX idx_vacancy_applications_unseen
    ON vacancy_applications (vacancy_id, seen_at)
    WHERE seen_at IS NULL;
