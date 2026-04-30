CREATE TABLE IF NOT EXISTS user_technologies (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    technology_id UUID NOT NULL REFERENCES technologies(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, technology_id)
);

CREATE TRIGGER update_user_technologies_updated_at
    BEFORE UPDATE ON user_technologies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS vacancy_technologies (
    vacancy_id UUID NOT NULL REFERENCES vacancies(id) ON DELETE CASCADE,
    technology_id UUID NOT NULL REFERENCES technologies(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (vacancy_id, technology_id)
);

CREATE TRIGGER update_vacancy_technologies_updated_at
    BEFORE UPDATE ON vacancy_technologies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
