# User Fanout Feature Documentation

## Overview

The `user_fanout` feature implements **Stage 2** of the 4-stage SQS pipeline. It performs a "fan-out" operation: for a given job, it fetches all users with AI prompts and enqueues a separate message for each user to the user analysis stage.

**Pipeline Stage**: Stage 2 - User Fanout  
**Queue**: `jobping-user-fanout` (input)  
**Next Queue**: `jobping-user-analysis` (output)

## File Structure

```
backend/internal/features/user_fanout/
├── handler/
│   └── sqs.go              # SQS event handler
└── service/
    └── fanout_service.go   # Core business logic
```

## Components

### Service - Fanout (`service/fanout_service.go`)

**Purpose**: Core business logic for Stage 2 - fanning out a job to all users.

**Key Method**:
```go
FanoutToUsers(ctx context.Context, jobID uuid.UUID) error
```

**Process Flow**:
1. **Fetch Users**: Retrieves all users with AI prompts from database
2. **Filter Users**: Skips users without AI prompts
3. **Enqueue Each User**: For each user, sends `{job_id, user_id}` to `user-analysis-queue`
4. **Logging**: Logs how many users were enqueued

**Dependencies**:
- `UserRepository` - Database operations
- SQS Client - For enqueueing to next stage

**Environment Variables**:
- `USER_ANALYSIS_QUEUE_URL` - Queue URL for next stage

**Error Handling**:
- If user fetch fails: returns error (will retry)
- If individual user enqueue fails: logs and continues (doesn't fail entire operation)
- If SQS not configured: logs and returns nil (no error)

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
3. **Call Service**: Calls `FanoutService.FanoutToUsers`
4. **Error Handling**: Logs errors but continues processing other messages

**Message Format** (Input):
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Message Format** (Output - sent to user-analysis-queue):
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Usage**: Used by `user_fanout_worker` Lambda function.

**Entry Point**: `cmd/workers/user_fanout/main.go`

---

## Data Flow

```
jobping-user-fanout queue
    ↓ (triggers Lambda)
user_fanout_worker Lambda
    ↓ (calls handler)
FanoutService.FanoutToUsers()
    ↓
1. Fetch all users with AI prompts from RDS
2. For each user:
   - Enqueue {job_id, user_id} to user-analysis-queue
    ↓
jobping-user-analysis queue
    ↓ (multiple messages, one per user)
```

---

## Database Operations

**Reads**:
- `users` table: `GetUsersWithPrompts()` - Fetches all users where `ai_prompt IS NOT NULL AND ai_prompt != ''`

**Writes**: None (read-only operation)

---

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `USER_ANALYSIS_QUEUE_URL` | No | SQS queue URL for next stage (skips if not set) |

---

## Error Handling

- **User fetch fails**: Returns error - message will be retried by SQS
- **Individual user enqueue fails**: Logs error, continues to next user - doesn't fail entire operation
- **SQS not configured**: Logs warning, returns nil - operation succeeds but no messages sent

**Design Decision**: Individual user enqueue failures don't fail the entire operation. This ensures that if one user fails, other users still get processed. Failed users can be retried by re-running the fanout.

---

## Scaling Considerations

**Fan-out Pattern**:
- One job message → N user messages (where N = number of users with prompts)
- If you have 100 users, one job creates 100 messages in the next queue
- This allows parallel processing of user analysis

**Performance**:
- Fetches all users in one query (efficient)
- Enqueues messages sequentially (could be optimized to batch)
- No AI calls in this stage (fast)

---

## Testing

**Local Testing**:
1. Start LocalStack with `jobping-user-fanout` and `jobping-user-analysis` queues
2. Ensure you have users with AI prompts in the database
3. Send test message:
   ```bash
   awslocal sqs send-message \
     --queue-url http://localhost:4566/000000000000/jobping-user-fanout \
     --message-body '{"job_id": "your-job-uuid"}'
   ```
4. Check `jobping-user-analysis` queue for messages (should have one per user)

**Production Testing**:
- Monitor CloudWatch logs for `user_fanout_worker` Lambda
- Check DLQ for failed messages
- Verify messages appear in `jobping-user-analysis` queue
- Count messages should match number of users with prompts

---

## Design Principles

1. **Single Responsibility**: Only handles fan-out, no AI analysis
2. **Resilient**: Individual failures don't fail entire operation
3. **Scalable**: Fan-out pattern allows parallel processing downstream
4. **Observable**: Logs number of users enqueued

---

## Example

**Input** (from job-analysis-queue):
```json
{
  "job_id": "job-123"
}
```

**Database Query**:
```sql
SELECT * FROM users WHERE ai_prompt IS NOT NULL AND ai_prompt != '';
-- Returns: [user-1, user-2, user-3]
```

**Output** (3 messages sent to user-analysis-queue):
```json
{"job_id": "job-123", "user_id": "user-1"}
{"job_id": "job-123", "user_id": "user-2"}
{"job_id": "job-123", "user_id": "user-3"}
```


