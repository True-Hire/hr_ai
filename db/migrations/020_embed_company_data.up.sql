-- Add company_data JSONB column to company_hrs
ALTER TABLE company_hrs ADD COLUMN IF NOT EXISTS company_data JSONB;

-- Add company_data JSONB column to vacancies
ALTER TABLE vacancies ADD COLUMN IF NOT EXISTS company_data JSONB;

-- Migrate existing company data into company_hrs.company_data
UPDATE company_hrs ch
SET company_data = (
    SELECT jsonb_build_object(
        'employee_count', COALESCE(c.employee_count, 0),
        'country', COALESCE(c.country, ''),
        'address', COALESCE(c.address, ''),
        'phone', COALESCE(c.phone, ''),
        'telegram', COALESCE(c.telegram, ''),
        'telegram_channel', COALESCE(c.telegram_channel, ''),
        'email', COALESCE(c.email, ''),
        'logo_url', COALESCE(c.logo_url, ''),
        'web_site', COALESCE(c.web_site, ''),
        'instagram', COALESCE(c.instagram, ''),
        'source_lang', COALESCE(c.source_lang, 'ru'),
        'texts', COALESCE((
            SELECT jsonb_agg(jsonb_build_object(
                'lang', ct.lang,
                'name', ct.name,
                'activity_type', ct.activity_type,
                'company_type', ct.company_type,
                'about', ct.about,
                'market', ct.market,
                'is_source', ct.is_source
            ) ORDER BY ct.lang)
            FROM company_texts ct WHERE ct.company_id = c.id
        ), '[]'::jsonb)
    )
    FROM companies c
    WHERE c.id = ch.company_id
)
WHERE ch.company_id IS NOT NULL;

-- Migrate existing company data into vacancies.company_data (snapshot from the company)
UPDATE vacancies v
SET company_data = (
    SELECT jsonb_build_object(
        'employee_count', COALESCE(c.employee_count, 0),
        'country', COALESCE(c.country, ''),
        'address', COALESCE(c.address, ''),
        'phone', COALESCE(c.phone, ''),
        'telegram', COALESCE(c.telegram, ''),
        'telegram_channel', COALESCE(c.telegram_channel, ''),
        'email', COALESCE(c.email, ''),
        'logo_url', COALESCE(c.logo_url, ''),
        'web_site', COALESCE(c.web_site, ''),
        'instagram', COALESCE(c.instagram, ''),
        'source_lang', COALESCE(c.source_lang, 'ru'),
        'texts', COALESCE((
            SELECT jsonb_agg(jsonb_build_object(
                'lang', ct.lang,
                'name', ct.name,
                'activity_type', ct.activity_type,
                'company_type', ct.company_type,
                'about', ct.about,
                'market', ct.market,
                'is_source', ct.is_source
            ) ORDER BY ct.lang)
            FROM company_texts ct WHERE ct.company_id = c.id
        ), '[]'::jsonb)
    )
    FROM companies c
    WHERE c.id = v.company_id
)
WHERE v.company_id IS NOT NULL;

-- Drop FK constraints and company_id columns
ALTER TABLE company_hrs DROP COLUMN IF EXISTS company_id;
ALTER TABLE vacancies DROP COLUMN IF EXISTS company_id;

-- Drop index on vacancies.company_id (if exists)
DROP INDEX IF EXISTS idx_vacancies_company_id;

-- Add GIN index on vacancies.company_data for company name search
CREATE INDEX IF NOT EXISTS idx_vacancies_company_data ON vacancies USING GIN (company_data);
