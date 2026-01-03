-- Add AI prompt and notification settings to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS ai_prompt TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS discord_webhook TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS notify_threshold INTEGER DEFAULT 70;

-- Add company research data to jobs (JSONB for flexible schema)
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS company_info JSONB;

-- User-job match results table
CREATE TABLE IF NOT EXISTS user_job_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    score INTEGER NOT NULL,
    analysis JSONB NOT NULL,
    notified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, job_id)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_user_job_matches_user_id ON user_job_matches(user_id);
CREATE INDEX IF NOT EXISTS idx_user_job_matches_job_id ON user_job_matches(job_id);
CREATE INDEX IF NOT EXISTS idx_user_job_matches_score ON user_job_matches(score DESC);
CREATE INDEX IF NOT EXISTS idx_user_job_matches_notified ON user_job_matches(notified) WHERE notified = FALSE;

