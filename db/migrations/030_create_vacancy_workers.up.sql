CREATE TABLE IF NOT EXISTS vacancy_workers (
    id UUID PRIMARY KEY,
    vacancy_id UUID NOT NULL REFERENCES vacancies(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TRIGGER update_vacancy_workers_updated_at
    BEFORE UPDATE ON vacancy_workers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
