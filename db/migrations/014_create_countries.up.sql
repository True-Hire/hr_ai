CREATE TABLE IF NOT EXISTS countries (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    short_code TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_countries_short_code ON countries(short_code);

CREATE TABLE IF NOT EXISTS country_texts (
    country_id UUID REFERENCES countries(id),
    lang TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    is_source BOOLEAN NOT NULL,
    model_version TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (country_id, lang)
);

ALTER TABLE vacancies ADD COLUMN country_id UUID REFERENCES countries(id);
CREATE INDEX idx_vacancies_country_id ON vacancies(country_id);
