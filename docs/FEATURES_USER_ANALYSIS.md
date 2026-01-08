# User Analysis Feature Documentation

## Overview

The `user_analysis` feature implements **Stage 3** of the 4-stage SQS pipeline. It performs per-user AI analysis to determine if a job matches a user's preferences. If a match is found (score >= threshold), it enqueues the user to the notification stage.

**Pipeline Stage**: Stage 3 - User Analysis  
**Queue**: `jobping-user-analysis` (input)  
**Next Queue**: `jobping-notification` (output, only if match found)

## File Structure

```
backend/internal/features/user_analysis/
├── handler/
│   └── sqs.go                    # SQS event handler
└── service/
    ├── ai_client.go             # AI client for user-job matching
    └── user_analysis_service.go # Core business logic
```

## Components

### Service - User Analysis (`service/user_analysis_service.go`)

**Purpose**: Core business logic for Stage 3 - per-user AI matching.

**Key Method**:
```go
AnalyzeUserMatch(ctx context.Context, jobID, userID uuid.UUID) error
```

**Process Flow**:
1. **Fetch Job**: Retrieves job from database by `job_id`
2. **Fetch User**: Retrieves user from database by `user_id`
3. **Validate**: Checks if user has AI prompt (skips if not)
4. **Check Existing Match**: 
   - If match already exists and not notified: enqueues to notification (if score >= threshold)
   - If match already exists and notified: returns (no-op)
5. **Run AI Matching** (if no existing match):
   - Calls AI client with job details and user prompt
   - Gets match score (0-100) and analysis
6. **Store Match**: Saves match to `user_job_matches` table
7. **Enqueue to Notification** (if score >= user's threshold):
   - Sends `{job_id, user_id}` to `notification-queue`

**Dependencies**:
- `JobRepository` - Database operations
- `UserRepository` - Database operations
- `UserJobMatchRepository` - Match storage
- `AIClient` - User-job matching (OpenAI)
- SQS Client - For enqueueing to next stage

**Environment Variables**:
- `NOTIFICATION_QUEUE_URL` - Queue URL for next stage

**Error Handling**:
- If job/user not found: logs and returns nil (no error)
- If user has no prompt: logs and returns nil (no error)
- If AI matching fails: returns error (will retry)
- If match save fails: returns error (will retry)
- If notification enqueue fails: returns error (will retry)

---

### Service - AI Client (`service/ai_client.go`)

**Purpose**: AI client interface for user-job matching using ChatGPT.

**Interface**:
```go
type AIClient interface {
    MatchJobToUser(ctx context.Context, job *JobMatchInput, userPrompt string) (*UserMatchResult, error)
}
```

**Input Types**:
```go
type JobMatchInput struct {
    Title       string
    Company     string
    Description string
    CompanyInfo map[string]interface{}  // From Stage 1 research
}
```

**Output Types**:
```go
type UserMatchResult struct {
    Score       int                    // 0-100
    Explanation string                 // 2-3 sentence explanation
    Pros        []string               // Reasons this is a good fit
    Cons        []string               // Reasons this might not be ideal
    Analysis    map[string]interface{} // Full analysis for storage
}
```

**Implementation**: `openAIClient`
- Uses OpenAI GPT-3.5-turbo API
- Returns mock data if `OPENAI_API_KEY` not set
- Analyzes job against user's preferences/ideal job description
- Returns structured match analysis

**Environment Variables**:
- `OPENAI_API_KEY` - OpenAI API key (optional, uses mock if not set)

**Usage**: Called by `UserAnalysisService` for each user-job pair.

---

### Handler - SQS (`handler/sqs.go`)

**Purpose**: SQS event handler for Lambda function.

**Key Method**:
```go
HandleSQSEvent(ctx context.Context, sqsEvent events.SQSEvent) error
```

**Process Flow**:
1. **Parse Messages**: Iterates through SQS event records
2. **Extract IDs**: Parses `{"job_id": "uuid", "user_id": "uuid"}` from message body
3. **Call Service**: Calls `UserAnalysisService.AnalyzeUserMatch`
4. **Error Handling**: Logs errors but continues processing other messages

**Message Format** (Input):
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Message Format** (Output - sent to notification-queue, only if match):
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Usage**: Used by `user_analysis_worker` Lambda function.

**Entry Point**: `cmd/workers/user_analysis/main.go`

---

## Data Flow

```
jobping-user-analysis queue
    ↓ (triggers Lambda, one message per user)
user_analysis_worker Lambda
    ↓ (calls handler)
UserAnalysisService.AnalyzeUserMatch()
    ↓
1. Fetch job from RDS
2. Fetch user from RDS
3. Check if match already exists
4. If not: Run AI matching (OpenAI)
5. Save match to user_job_matches table
6. If score >= threshold: Enqueue to notification-queue
    ↓ (only if match found)
jobping-notification queue
```

---

## Database Operations

**Reads**:
- `jobs` table: `GetByID(job_id)`
- `users` table: `GetUserByID(user_id)`
- `user_job_matches` table: `GetByUserAndJob(user_id, job_id)`

**Writes**:
- `user_job_matches` table: `Create(match)` - Stores:
  - `user_id`, `job_id`
  - `score` (0-100)
  - `analysis` (JSONB with full AI analysis)
  - `notified` (false initially)

---

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `NOTIFICATION_QUEUE_URL` | No | SQS queue URL for next stage (skips if not set) |
| `OPENAI_API_KEY` | No | OpenAI API key (uses mock if not set) |

---

## Error Handling

- **Job/user not found**: Logs warning, returns nil - message is consumed
- **User has no prompt**: Logs warning, returns nil - message is consumed
- **AI matching fails**: Returns error - message will be retried
- **Match save fails**: Returns error - message will be retried
- **Notification enqueue fails**: Returns error - message will be retried

**Idempotency**: 
- Checks for existing match before running AI
- If match exists and not notified, still enqueues to notification (if score >= threshold)
- Prevents duplicate AI calls for same user-job pair

---

## Matching Logic

**Match Score Threshold**:
- Each user has a `notify_threshold` (default: 70)
- Only matches with `score >= notify_threshold` are enqueued to notification
- Matches below threshold are still saved but not notified

**AI Analysis Includes**:
- Match score (0-100)
- Explanation of why it matches/doesn't match
- Pros (reasons it's a good fit)
- Cons (reasons it might not be ideal)
- Key match factors (specific factors from user preferences)

**Company Info Usage**:
- Uses company research from Stage 1 (`company_info` field)
- Provides context to AI for better matching

---

## Testing

**Local Testing**:
1. Start LocalStack with `jobping-user-analysis` and `jobping-notification` queues
2. Ensure you have:
   - A job in the database
   - A user with AI prompt in the database
3. Send test message:
   ```bash
   awslocal sqs send-message \
     --queue-url http://localhost:4566/000000000000/jobping-user-analysis \
     --message-body '{"job_id": "job-uuid", "user_id": "user-uuid"}'
   ```
4. Check logs for AI analysis
5. Check `user_job_matches` table for match record
6. Check `jobping-notification` queue (should have message if score >= threshold)

**Production Testing**:
- Monitor CloudWatch logs for `user_analysis_worker` Lambda
- Check DLQ for failed messages
- Verify matches in `user_job_matches` table
- Verify messages in `jobping-notification` queue (only for matches)

---

## Design Principles

1. **Single Responsibility**: One AI call per message (one user-job pair)
2. **Idempotent**: Can be safely retried (checks for existing match)
3. **Resilient**: Handles missing data gracefully
4. **Observable**: Logs all operations and match scores

---

## Cost Considerations

**AI API Calls**:
- One OpenAI API call per user-job pair
- If 100 users and 1 job = 100 API calls
- If 100 users and 10 jobs = 1,000 API calls
- Cost scales with number of users × number of jobs

**Optimization Opportunities**:
- Batch processing (not implemented)
- Caching company info (already done in Stage 1)
- Skipping low-probability matches (not implemented)


