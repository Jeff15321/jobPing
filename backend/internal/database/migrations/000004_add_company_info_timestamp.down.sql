-- Remove company_info timestamp
DROP INDEX IF EXISTS idx_jobs_company_info_updated;
ALTER TABLE jobs DROP COLUMN IF EXISTS company_info_updated_at;


