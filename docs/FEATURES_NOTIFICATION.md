# Notification Feature Documentation

## Overview

The `notification` feature implements **Stage 4** of the 4-stage SQS pipeline. It creates notification events for users when they match with jobs. For testing purposes, notifications are stored in the database and can be retrieved via HTTP API. In production, this would send actual notifications (email/push/webhook).

**Pipeline Stage**: Stage 4 - Notification  
**Queue**: `jobping-notification` (input)

## File Structure

```
backend/internal/features/notification/
├── handler/
│   ├── http.go                    # HTTP handler for fetching notifications
│   └── sqs.go                     # SQS event handler
├── repository/
│   └── notification_repository.go # Database operations
└── service/
    └── notification_service.go    # Core business logic
```

## Components

### Service - Notification (`service/notification_service.go`)

**Purpose**: Core business logic for Stage 4 - creating notification events.

**Key Methods**:

1. **SendNotification**:
```go
SendNotification(ctx context.Context, jobID, userID uuid.UUID) error
```

**Process Flow**:
1. **Fetch User**: Retrieves user from database by `user_id`
2. **Fetch Job**: Retrieves job from database by `job_id`
3. **Fetch Match**: Retrieves match from `user_job_matches` table
4. **Create Notification**: Creates notification record in `notifications` table with:
   - User, job, and match IDs
   - Job title, company, URL
   - Matching score
   - AI analysis (from match record)
5. **Mark Match as Notified**: Updates `user_job_matches.notified = true`

**Dependencies**:
- `JobRepository` - Database operations
- `UserRepository` - Database operations
- `UserJobMatchRepository` - Match operations
- `NotificationRepository` - Notification storage

**Error Handling**:
- If user/job/match not found: logs and returns nil (no error)
- If notification creation fails: returns error (will retry)
- If marking notified fails: logs error but doesn't fail operation

2. **GetNotifications**:
```go
GetNotifications(ctx context.Context, userID *uuid.UUID, limit int) ([]Notification, error)
```

**Purpose**: Retrieves notifications for display (testing/development).

**Process Flow**:
- If `userID` provided: fetches notifications for that user
- If `userID` is nil: fetches all notifications
- Returns notifications ordered by `created_at DESC`

---

### Repository - Notification (`repository/notification_repository.go`)

**Purpose**: Database operations for notifications.

**Interface**:
```go
type NotificationRepository interface {
    Create(ctx context.Context, notification *Notification) error
    GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]Notification, error)
    GetAll(ctx context.Context, limit int) ([]Notification, error)
}
```

**Model**:
```go
type Notification struct {
    ID            uuid.UUID
    UserID        uuid.UUID
    JobID         uuid.UUID
    MatchID       uuid.UUID
    JobTitle      string
    Company       string
    JobURL        string
    MatchingScore int
    AIAnalysis    map[string]interface{}  // Full AI analysis from match
    CreatedAt     time.Time
}
```

**Database Table**: `notifications`

**Key Methods**:
- `Create` - Inserts notification record
- `GetByUserID` - Fetches notifications for a user
- `GetAll` - Fetches all notifications (for testing)

**Indexes**:
- `idx_notifications_user_id` - On `user_id`
- `idx_notifications_created_at` - On `created_at DESC`

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
3. **Call Service**: Calls `NotificationService.SendNotification`
4. **Error Handling**: Logs errors but continues processing other messages

**Message Format** (Input):
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Usage**: Used by `notifier_worker` Lambda function.

**Entry Point**: `cmd/workers/notifier/main.go`

---

### Handler - HTTP (`handler/http.go`)

**Purpose**: HTTP handler for fetching notifications (testing/development).

**Endpoint**: `GET /api/notifications`

**Query Parameters**:
- `user_id` (optional) - Filter by user ID
- `limit` (optional) - Limit results (default: 20)

**Response Format**:
```json
{
  "notifications": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "job_id": "uuid",
      "match_id": "uuid",
      "job_title": "Software Engineer",
      "company": "TechCorp",
      "job_url": "https://...",
      "matching_score": 85,
      "ai_analysis": {
        "score": 85,
        "explanation": "...",
        "pros": [...],
        "cons": [...]
      },
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Usage**: Used by frontend to display notifications for testing.

**Entry Point**: `jobs_api` Lambda (via `internal/server/router.go`)

---

## Data Flow

```
jobping-notification queue
    ↓ (triggers Lambda)
notifier_worker Lambda
    ↓ (calls handler)
NotificationService.SendNotification()
    ↓
1. Fetch user from RDS
2. Fetch job from RDS
3. Fetch match from user_job_matches table
4. Create notification in notifications table
5. Mark match as notified
    ↓
Frontend
    ↓ (GET /api/notifications)
NotificationService.GetNotifications()
    ↓
Returns notifications with AI analysis
```

---

## Database Operations

**Reads**:
- `users` table: `GetUserByID(user_id)`
- `jobs` table: `GetByID(job_id)`
- `user_job_matches` table: `GetByUserAndJob(user_id, job_id)`
- `notifications` table: `GetByUserID(user_id, limit)` or `GetAll(limit)`

**Writes**:
- `notifications` table: `Create(notification)` - Stores notification event
- `user_job_matches` table: `MarkNotified(match_id)` - Updates `notified = true`

---

## Database Schema

**Table**: `notifications`

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    job_id UUID NOT NULL REFERENCES jobs(id),
    match_id UUID NOT NULL REFERENCES user_job_matches(id),
    job_title VARCHAR(500) NOT NULL,
    company VARCHAR(255) NOT NULL,
    job_url TEXT NOT NULL,
    matching_score INTEGER NOT NULL,
    ai_analysis JSONB NOT NULL,  -- Full AI analysis from match
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
```

**Migration**: `000005_add_notifications_table.up.sql`

---

## Environment Variables

None required (all operations use database only).

---

## Error Handling

- **User/job/match not found**: Logs warning, returns nil - message is consumed
- **Notification creation fails**: Returns error - message will be retried
- **Mark notified fails**: Logs error but doesn't fail operation - notification still created

**Idempotency**:
- If match is already notified, creating another notification is safe (just creates duplicate record)
- Consider adding check to prevent duplicate notifications (not implemented)

---

## Testing

**Local Testing**:
1. Start LocalStack with `jobping-notification` queue
2. Ensure you have:
   - A job in the database
   - A user in the database
   - A match in `user_job_matches` table
3. Send test message:
   ```bash
   awslocal sqs send-message \
     --queue-url http://localhost:4566/000000000000/jobping-notification \
     --message-body '{"job_id": "job-uuid", "user_id": "user-uuid"}'
   ```
4. Check `notifications` table for new record
5. Check `user_job_matches` table - `notified` should be `true`
6. Fetch via HTTP: `GET http://localhost:8080/api/notifications`

**Production Testing**:
- Monitor CloudWatch logs for `notifier_worker` Lambda
- Check DLQ for failed messages
- Verify notifications in `notifications` table
- Test HTTP endpoint: `GET /api/notifications`

---

## Future Enhancements

**Current Implementation** (Testing):
- Stores notifications in database
- HTTP endpoint for fetching notifications
- Frontend displays notifications with AI analysis

**Production Implementation** (To be added):
- Email notifications
- Push notifications
- Webhook notifications (Discord, Slack, etc.)
- Notification preferences per user
- Notification delivery status tracking

**Design Note**: The current implementation stores notifications for testing. In production, you would:
1. Keep the notification record (for history)
2. Send actual notification (email/push/webhook)
3. Track delivery status

---

## Design Principles

1. **Single Responsibility**: Only handles notification creation and retrieval
2. **Idempotent**: Can be safely retried (creates notification record)
3. **Observable**: Logs all operations
4. **Testable**: Stores notifications in database for testing

---

## AI Analysis in Notifications

**Source**: AI analysis comes from `user_job_matches.analysis` field (created in Stage 3).

**Content**:
- Match score (0-100)
- Explanation of match
- Pros (reasons it's a good fit)
- Cons (reasons it might not be ideal)
- Key match factors

**Usage**: Displayed to user in frontend to help them understand why they were matched with the job.


