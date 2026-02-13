CREATE TABLE IF NOT EXISTS vacancies (
    id UUID PRIMARY KEY,
    hr_id UUID NOT NULL REFERENCES company_hrs(id),
    company_id UUID NOT NULL REFERENCES companies(id),
    salary_min INT,
    salary_max INT,
    salary_currency TEXT NOT NULL DEFAULT 'USD',
    experience_min INT,
    experience_max INT,
    format TEXT NOT NULL DEFAULT 'office',
    schedule TEXT NOT NULL DEFAULT 'full-time',
    phone TEXT,
    telegram TEXT,
    email TEXT,
    address TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    source_lang TEXT NOT NULL DEFAULT 'ru',
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_vacancies_company_id ON vacancies(company_id);
CREATE INDEX idx_vacancies_hr_id ON vacancies(hr_id);
CREATE INDEX idx_vacancies_status ON vacancies(status);

CREATE TABLE IF NOT EXISTS vacancy_texts (
    vacancy_id UUID REFERENCES vacancies(id),
    lang TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    responsibilities TEXT NOT NULL DEFAULT '',
    requirements TEXT NOT NULL DEFAULT '',
    benefits TEXT NOT NULL DEFAULT '',
    is_source BOOLEAN NOT NULL,
    model_version TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (vacancy_id, lang)
);

CREATE TABLE IF NOT EXISTS vacancy_skills (
    vacancy_id UUID NOT NULL REFERENCES vacancies(id),
    skill_id UUID NOT NULL REFERENCES skills(id),
    PRIMARY KEY (vacancy_id, skill_id)
);
