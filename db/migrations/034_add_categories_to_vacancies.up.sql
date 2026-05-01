-- Add category columns to vacancies table
ALTER TABLE vacancies ADD COLUMN IF NOT EXISTS main_category_id UUID REFERENCES main_category(id) ON DELETE SET NULL;
ALTER TABLE vacancies ADD COLUMN IF NOT EXISTS sub_category_id UUID REFERENCES sub_category(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_vacancies_main_category_id ON vacancies(main_category_id);
CREATE INDEX IF NOT EXISTS idx_vacancies_sub_category_id ON vacancies(sub_category_id);
