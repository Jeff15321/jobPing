-- Jobs table for storing scraped and AI-analyzed jobs
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(500) NOT NULL,
    company VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    job_url TEXT NOT NULL UNIQUE,
    description TEXT,
    job_type VARCHAR(50),
    is_remote BOOLEAN DEFAULT FALSE,
    min_salary DECIMAL(12, 2),
    max_salary DECIMAL(12, 2),
    date_posted VARCHAR(50),
    ai_score INTEGER,
    ai_analysis TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_ai_score ON jobs(ai_score DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_jobs_company ON jobs(company);



