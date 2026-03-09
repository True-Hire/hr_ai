CREATE TABLE hr_saved_users (
    hr_id   UUID NOT NULL REFERENCES company_hrs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    note    TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (hr_id, user_id)
);

CREATE INDEX idx_hr_saved_users_hr_id ON hr_saved_users (hr_id, created_at DESC);
