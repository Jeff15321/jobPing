-- Add timestamp for company_info updates
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS company_info_updated_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_jobs_company_info_updated ON jobs(company_info_updated_at);


