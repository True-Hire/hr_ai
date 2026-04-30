ALTER TABLE skills ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT now();

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_skills_updated_at') THEN
        CREATE TRIGGER update_skills_updated_at
        BEFORE UPDATE ON skills
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
