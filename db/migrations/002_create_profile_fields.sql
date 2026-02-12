CREATE TABLE IF NOT EXISTS profile_fields (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    field_name TEXT NOT NULL,
    source_lang TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    UNIQUE (user_id, field_name)
);
