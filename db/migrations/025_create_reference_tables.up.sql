-- Company reference for prestige/engineering quality scoring
CREATE TABLE IF NOT EXISTS company_references (
    company_name TEXT PRIMARY KEY,
    normalized_name TEXT UNIQUE NOT NULL,
    prestige_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    engineering_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    scale_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    hiring_bar_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    category TEXT NOT NULL DEFAULT 'unknown',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- University/institution reference for education quality scoring
CREATE TABLE IF NOT EXISTS university_references (
    institution_name TEXT PRIMARY KEY,
    normalized_name TEXT UNIQUE NOT NULL,
    education_score NUMERIC(6,3) NOT NULL DEFAULT 0,
    category TEXT NOT NULL DEFAULT 'unknown',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed well-known companies
INSERT INTO company_references (company_name, normalized_name, prestige_score, engineering_score, scale_score, hiring_bar_score, category) VALUES
-- Global tech giants
('Google', 'google', 0.98, 0.98, 0.98, 0.98, 'big_tech'),
('Meta', 'meta', 0.95, 0.95, 0.95, 0.95, 'big_tech'),
('Apple', 'apple', 0.95, 0.93, 0.95, 0.95, 'big_tech'),
('Amazon', 'amazon', 0.93, 0.90, 0.98, 0.90, 'big_tech'),
('Microsoft', 'microsoft', 0.93, 0.90, 0.97, 0.90, 'big_tech'),
('Yandex', 'yandex', 0.90, 0.92, 0.90, 0.90, 'big_tech'),
('EPAM', 'epam', 0.75, 0.80, 0.85, 0.75, 'outsource_top'),
('Revolut', 'revolut', 0.85, 0.88, 0.80, 0.85, 'fintech'),
('Stripe', 'stripe', 0.90, 0.92, 0.80, 0.92, 'fintech'),
-- Uzbekistan companies
('Payme', 'payme', 0.70, 0.72, 0.65, 0.68, 'uz_fintech'),
('Click', 'click', 0.68, 0.70, 0.65, 0.65, 'uz_fintech'),
('Uzum', 'uzum', 0.72, 0.74, 0.70, 0.70, 'uz_ecommerce'),
('MyTaxi', 'mytaxi', 0.65, 0.68, 0.60, 0.62, 'uz_tech'),
('Billz', 'billz', 0.60, 0.65, 0.50, 0.60, 'uz_tech'),
('Humans', 'humans', 0.62, 0.64, 0.55, 0.60, 'uz_tech'),
('Apelsin', 'apelsin', 0.58, 0.60, 0.50, 0.55, 'uz_fintech'),
('OsonTaxi', 'osontaxi', 0.55, 0.55, 0.45, 0.50, 'uz_tech'),
('Korzinka', 'korzinka', 0.50, 0.45, 0.60, 0.45, 'uz_retail'),
('Artel', 'artel', 0.55, 0.50, 0.65, 0.50, 'uz_manufacturing'),
('UzAuto', 'uzauto', 0.55, 0.45, 0.70, 0.45, 'uz_manufacturing'),
('Ucell', 'ucell', 0.55, 0.50, 0.60, 0.50, 'uz_telecom'),
('Beeline Uzbekistan', 'beeline_uz', 0.58, 0.55, 0.65, 0.55, 'uz_telecom'),
('IT Park', 'it_park', 0.60, 0.55, 0.50, 0.55, 'uz_gov_tech'),
('Mediapark', 'mediapark', 0.55, 0.58, 0.50, 0.55, 'uz_media'),
('Kapitalbank', 'kapitalbank', 0.58, 0.55, 0.60, 0.55, 'uz_banking'),
('Fido Biznes', 'fido_biznes', 0.55, 0.52, 0.50, 0.50, 'uz_fintech'),
-- CIS / International
('Kaspersky', 'kaspersky', 0.82, 0.85, 0.80, 0.82, 'security'),
('JetBrains', 'jetbrains', 0.85, 0.90, 0.70, 0.88, 'dev_tools'),
('Tinkoff', 'tinkoff', 0.82, 0.85, 0.80, 0.82, 'fintech'),
('VK', 'vk', 0.78, 0.80, 0.82, 0.78, 'big_tech_cis')
ON CONFLICT (company_name) DO NOTHING;

-- Seed universities
INSERT INTO university_references (institution_name, normalized_name, education_score, category) VALUES
-- Uzbekistan
('TUIT', 'tuit', 0.65, 'uz_university'),
('Ташкентский университет информационных технологий', 'tuit_full', 0.65, 'uz_university'),
('WIUT', 'wiut', 0.72, 'uz_international'),
('Westminster International University in Tashkent', 'wiut_full', 0.72, 'uz_international'),
('Inha University in Tashkent', 'inha_tashkent', 0.70, 'uz_international'),
('MDIST', 'mdist', 0.60, 'uz_university'),
('Samarkand State University', 'samarkand_state', 0.55, 'uz_university'),
('Fergana State University', 'fergana_state', 0.50, 'uz_university'),
('Bukhara State University', 'bukhara_state', 0.50, 'uz_university'),
('Buchara Davlat Universiteti', 'bukhara_davlat', 0.50, 'uz_university'),
('Namangan State University', 'namangan_state', 0.48, 'uz_university'),
('Andijan State University', 'andijan_state', 0.48, 'uz_university'),
('Nukus State University', 'nukus_state', 0.45, 'uz_university'),
-- Bootcamps
('PDP Academy', 'pdp_academy', 0.55, 'bootcamp'),
('Najot Talim', 'najot_talim', 0.52, 'bootcamp'),
-- International top
('MIT', 'mit', 0.98, 'world_top'),
('Stanford', 'stanford', 0.98, 'world_top'),
('HSE', 'hse', 0.78, 'cis_top'),
('ITMO', 'itmo', 0.80, 'cis_top'),
('MSU', 'msu', 0.82, 'cis_top'),
('MIPT', 'mipt', 0.85, 'cis_top')
ON CONFLICT (institution_name) DO NOTHING;
