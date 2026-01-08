-- Notifications table for storing notification events (for testing)
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    match_id UUID NOT NULL REFERENCES user_job_matches(id) ON DELETE CASCADE,
    job_title VARCHAR(500) NOT NULL,
    company VARCHAR(255) NOT NULL,
    job_url TEXT NOT NULL,
    matching_score INTEGER NOT NULL,
    ai_analysis JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);


