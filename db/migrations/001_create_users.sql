CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    phone TEXT UNIQUE,
    email TEXT UNIQUE,
    profile_pic_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
