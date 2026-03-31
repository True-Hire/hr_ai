CREATE TABLE IF NOT EXISTS candidate_search_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

    -- Normalized fields
    primary_role TEXT NOT NULL DEFAULT '',
    role_family TEXT NOT NULL DEFAULT '',
    seniority TEXT NOT NULL DEFAULT '',
    total_experience_months INT NOT NULL DEFAULT 0,
    highest_education_level TEXT NOT NULL DEFAULT '',

    -- Array fields
    skills TEXT[] DEFAULT '{}',
    industries TEXT[] DEFAULT '{}',
    project_domains TEXT[] DEFAULT '{}',
    company_names TEXT[] DEFAULT '{}',
    known_languages TEXT[] DEFAULT '{}',
    education_fields TEXT[] DEFAULT '{}',
    universities TEXT[] DEFAULT '{}',

    -- Location
    location_city TEXT NOT NULL DEFAULT '',
    location_country TEXT NOT NULL DEFAULT '',
    willing_to_relocate BOOLEAN NOT NULL DEFAULT false,

    -- Role relevance scores (0.000 - 1.000)
    backend_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    frontend_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    mobile_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    data_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    qa_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    pm_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    devops_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    design_score NUMERIC(6,3) NOT NULL DEFAULT 0,

    -- Capability/bonus signals
    devops_support_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    client_communication_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    project_management_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    ownership_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    leadership_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    mentoring_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    startup_adaptability_score NUMERIC(6,3) NOT NULL DEFAULT 0,

    -- Market strength signals
    company_prestige_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    engineering_environment_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    internship_quality_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    education_quality_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    competition_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    open_source_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    growth_trajectory_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    project_complexity_score NUMERIC(6,3) NOT NULL DEFAULT 0,

    -- Aggregated strength scores
    overall_strength_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    backend_strength_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    frontend_strength_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    data_strength_score NUMERIC(6,3) NOT NULL DEFAULT 0,

    -- Search
    search_text TEXT NOT NULL DEFAULT '',
    search_tsv TSVECTOR,

    -- Debug / explainability
    scoring_factors JSONB DEFAULT '{}',
    parsed_entities JSONB DEFAULT '{}',

    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_csp_search_tsv ON candidate_search_profiles USING GIN (search_tsv);
CREATE INDEX idx_csp_primary_role ON candidate_search_profiles (primary_role);
CREATE INDEX idx_csp_role_family ON candidate_search_profiles (role_family);
CREATE INDEX idx_csp_seniority ON candidate_search_profiles (seniority);
CREATE INDEX idx_csp_experience ON candidate_search_profiles (total_experience_months);
CREATE INDEX idx_csp_location_city ON candidate_search_profiles (location_city);
CREATE INDEX idx_csp_location_country ON candidate_search_profiles (location_country);
CREATE INDEX idx_csp_skills ON candidate_search_profiles USING GIN (skills);
CREATE INDEX idx_csp_industries ON candidate_search_profiles USING GIN (industries);
CREATE INDEX idx_csp_overall_strength ON candidate_search_profiles (overall_strength_score DESC);
