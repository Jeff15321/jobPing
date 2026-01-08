-- Drop user_job_matches table
DROP TABLE IF EXISTS user_job_matches;

-- Remove columns from jobs
ALTER TABLE jobs DROP COLUMN IF EXISTS company_info;

-- Remove columns from users
ALTER TABLE users DROP COLUMN IF EXISTS notify_threshold;
ALTER TABLE users DROP COLUMN IF EXISTS discord_webhook;
ALTER TABLE users DROP COLUMN IF EXISTS ai_prompt;


