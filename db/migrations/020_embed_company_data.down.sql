-- Drop GIN index
DROP INDEX IF EXISTS idx_vacancies_company_data;

-- Re-add company_id columns
ALTER TABLE company_hrs ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES companies(id);
ALTER TABLE vacancies ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES companies(id);

-- Re-create index
CREATE INDEX IF NOT EXISTS idx_vacancies_company_id ON vacancies(company_id);

-- Drop company_data columns
ALTER TABLE company_hrs DROP COLUMN IF EXISTS company_data;
ALTER TABLE vacancies DROP COLUMN IF EXISTS company_data;
