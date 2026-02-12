CREATE TABLE IF NOT EXISTS profile_field_texts (
    profile_field_id UUID REFERENCES profile_fields(id),
    lang TEXT NOT NULL,
    content TEXT NOT NULL,
    is_source BOOLEAN NOT NULL,
    model_version TEXT,
    updated_at TIMESTAMP NOT NULL,

    PRIMARY KEY (profile_field_id, lang)
);
