CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY,
    employee_count INT,
    country TEXT,
    address TEXT,
    phone TEXT,
    telegram TEXT,
    telegram_channel TEXT,
    email TEXT,
    logo_url TEXT,
    web_site TEXT,
    instagram TEXT,
    source_lang TEXT NOT NULL DEFAULT 'ru',
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS company_texts (
    company_id UUID REFERENCES companies(id),
    lang TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    activity_type TEXT NOT NULL DEFAULT '',
    company_type TEXT NOT NULL DEFAULT '',
    about TEXT NOT NULL DEFAULT '',
    market TEXT NOT NULL DEFAULT '',
    is_source BOOLEAN NOT NULL,
    model_version TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),

    PRIMARY KEY (company_id, lang)
);

ALTER TABLE company_hrs
    ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES companies(id),
    DROP COLUMN IF EXISTS company_name,
    DROP COLUMN IF EXISTS activity_type,
    DROP COLUMN IF EXISTS company_type,
    DROP COLUMN IF EXISTS employee_count,
    DROP COLUMN IF EXISTS country,
    DROP COLUMN IF EXISTS market,
    DROP COLUMN IF EXISTS web_site,
    DROP COLUMN IF EXISTS about,
    DROP COLUMN IF EXISTS logo_url,
    DROP COLUMN IF EXISTS instagram;
