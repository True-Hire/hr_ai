CREATE TABLE IF NOT EXISTS vacancy_applications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    vacancy_id UUID NOT NULL REFERENCES vacancies(id),
    status TEXT NOT NULL DEFAULT 'pending',
    cover_letter TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (user_id, vacancy_id)
);

CREATE INDEX idx_vacancy_applications_user_id ON vacancy_applications(user_id);
CREATE INDEX idx_vacancy_applications_vacancy_id ON vacancy_applications(vacancy_id);
