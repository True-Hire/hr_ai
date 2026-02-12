CREATE TABLE IF NOT EXISTS item_texts (
    item_id UUID NOT NULL,
    item_type TEXT NOT NULL,
    lang TEXT NOT NULL,
    description TEXT NOT NULL,
    is_source BOOLEAN NOT NULL DEFAULT false,
    model_version TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (item_id, item_type, lang)
);
