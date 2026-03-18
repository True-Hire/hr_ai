CREATE TABLE IF NOT EXISTS search_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hr_id UUID NOT NULL,
    query_text TEXT NOT NULL,
    parsed_query JSONB NOT NULL DEFAULT '{}',
    filters JSONB NOT NULL DEFAULT '{}',
    total_results INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'ready',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '1 hour')
);

CREATE TABLE IF NOT EXISTS search_session_results (
    search_id UUID NOT NULL REFERENCES search_sessions(id) ON DELETE CASCADE,
    rank INT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    final_score NUMERIC(10,4) NOT NULL,
    score_breakdown JSONB NOT NULL DEFAULT '{}',
    PRIMARY KEY (search_id, rank),
    UNIQUE (search_id, user_id)
);

CREATE INDEX idx_ssr_search_rank ON search_session_results (search_id, rank);
CREATE INDEX idx_search_sessions_hr ON search_sessions (hr_id);
CREATE INDEX idx_search_sessions_expires ON search_sessions (expires_at);
