CREATE TABLE IF NOT EXISTS company_hrs (
    id UUID PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    patronymic TEXT,
    phone TEXT UNIQUE,
    telegram TEXT UNIQUE,
    telegram_id TEXT UNIQUE,
    email TEXT UNIQUE,
    position TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    password_hash TEXT,
    company_name TEXT,
    activity_type TEXT,
    company_type TEXT,
    employee_count INT,
    country TEXT,
    market TEXT,
    web_site TEXT,
    about TEXT,
    logo_url TEXT,
    instagram TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS hr_sessions (
    id UUID PRIMARY KEY,
    hr_id UUID NOT NULL REFERENCES company_hrs(id),
    device_id TEXT NOT NULL,
    refresh_token_hash TEXT NOT NULL,
    fcm_token TEXT,
    ip_address TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_hr_sessions_hr_id ON hr_sessions(hr_id);
CREATE INDEX IF NOT EXISTS idx_hr_sessions_device_id ON hr_sessions(device_id);
