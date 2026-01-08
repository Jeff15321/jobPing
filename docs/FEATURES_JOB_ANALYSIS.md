# Job Analysis Feature Documentation

## Overview

The `job_analysis` feature implements **Stage 1** of the 4-stage SQS pipeline. It handles company research for new jobs and enqueues jobs to the user fanout stage.

**Pipeline Stage**: Stage 1 - Job Analysis  
**Queue**: `jobping-job-analysis` (input)  
**Next Queue**: `jobping-user-fanout` (output)

## File Structure

```
backend/internal/features/job_analysis/
├── handler/
│   └── sqs.go                    # SQS event handler
└── service/
    ├── ai_client.go             # AI client for company research
    └── job_analysis_service.go  # Core business logic
```

## Components

### Service - Job Analysis (`service/job_analysis_service.go`)

**Purpose**: Core business logic for Stage 1 - company research and fanout enqueueing.

**Key Method**:
```go
AnalyzeJob(ctx context.Context, jobID uuid.UUID) error
```

**Process Flow**:
1. **Fetch Job**: Retrieves job from database by `job_id`
2. **Check Freshness**: Checks if company info exists and is less than 6 months old
3. **Research Company** (if not fresh):
   - Calls AI client to research company
   - Saves company info to database
   - Updates `company_info_updated_at` timestamp
4. **Enqueue to Fanout**: Sends `job_id` to `user-fanout-queue`

**Dependencies**:
- `JobRepository` - Database operations
- `AIClient` - Company research (OpenAI)
- SQS Client - For enqueueing to next stage

**Environment Variables**:
- `USER_FANOUT_QUEUE_URL` - Queue URL for next stage

**Error Handling**:
- If job not found: logs and returns nil (no error)
- If company research fails: logs and continues to fanout
- If fanout enqueue fails: returns error (will retry)

---

### Service - AI Client (`service/ai_client.go`)

**Purpose**: AI client interface for company research using ChatGPT.

**Interface**:
```go
type AIClient interface {
    ResearchCompany(ctx context.Context, company, title, description string) (map[string]interface{}, error)
}
```

**Implementation**: `openAIClient`
- Uses OpenAI GPT-3.5-turbo API
- Returns mock data if `OPENAI_API_KEY` not set
- Returns JSON object with company information:
  - `company_size`, `industry`, `culture`, `funding`
  - `notable_info`, `tech_stack`, `work_life_balance`
  - `red_flags`, `green_flags`

**Environment Variables**:
- `OPENAI_API_KEY` - OpenAI API key (optional, uses mock if not set)

**Usage**: Called by `JobAnalysisService` when company info is stale or missing.

---

### Handler - SQS (`handler/sqs.go`)

**Purpose**: SQS event handler for Lambda function.

**Key Method**:
```go
HandleSQSEvent(ctx context.Context, sqsEvent events.SQSEvent) error
```

**Process Flow**:
1. **Parse Messages**: Iterates through SQS event records
2. **Extract Job ID**: Parses `{"job_id": "uuid"}` from message body
3. **Call Service**: Calls `JobAnalysisService.AnalyzeJob`
4. **Error Handling**: Logs errors but continues processing other messages

**Message Format**:
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Usage**: Used by `job_analysis_worker` Lambda function.

**Entry Point**: `cmd/workers/job_analysis/main.go`

---

## Data Flow

```
Python jobspy_fetcher
    ↓ (creates job in RDS)
    ↓ (sends job_id to SQS)
jobping-job-analysis queue
    ↓ (triggers Lambda)
job_analysis_worker Lambda
    ↓ (calls handler)
JobAnalysisService.AnalyzeJob()
    ↓
1. Fetch job from RDS
2. Check company_info_updated_at
3. If stale: Research company (OpenAI)
4. Save company info to RDS
5. Enqueue job_id to user-fanout-queue
    ↓
jobping-user-fanout queue
```

---

## Database Operations

**Reads**:
- `jobs` table: `GetByID(job_id)`
- `jobs` table: `IsCompanyInfoFresh(job_id)` - Checks if `company_info_updated_at` < 6 months

**Writes**:
- `jobs` table: `UpdateCompanyInfo(job_id, companyInfo)` - Updates `company_info` and `company_info_updated_at`

---

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `USER_FANOUT_QUEUE_URL` | No | SQS queue URL for next stage (skips if not set) |
| `OPENAI_API_KEY` | No | OpenAI API key (uses mock if not set) |

---

## Error Handling

- **Job not found**: Logs warning, returns nil (no error) - message is consumed
- **Company research fails**: Logs error, continues to fanout - job still processed
- **Fanout enqueue fails**: Returns error - message will be retried by SQS
- **Database errors**: Returns error - message will be retried

---

## Testing

**Local Testing**:
1. Start LocalStack with `jobping-job-analysis` queue
2. Send test message:
   ```bash
   awslocal sqs send-message \
     --queue-url http://localhost:4566/000000000000/jobping-job-analysis \
     --message-body '{"job_id": "your-job-uuid"}'
   ```
3. Check logs for processing

**Production Testing**:
- Monitor CloudWatch logs for `job_analysis_worker` Lambda
- Check DLQ for failed messages
- Verify company info is updated in RDS

---

## Design Principles

1. **Single Responsibility**: Only handles company research and fanout enqueueing
2. **Idempotent**: Can be safely retried (checks freshness before research)
3. **Resilient**: Continues to fanout even if company research fails
4. **Observable**: Logs all operations for debugging


