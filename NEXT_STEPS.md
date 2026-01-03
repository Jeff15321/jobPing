# Implementation Status

## Completed Features

### User AI Prompt Configuration
- Users can set a custom AI prompt via `PUT /api/users/me/prompt`
- Users can configure Discord webhook via `PUT /api/users/me/discord`
- Users can set notification threshold (0-100) via `PUT /api/users/me/threshold`
- View job matches via `GET /api/users/me/matches`

### Job Enrichment
- Each new job is analyzed with OpenAI to research the company
- Company info stored as JSONB in the `company_info` column
- Includes: company size, culture, funding, tech stack, red/green flags

### Per-User Job Matching
- Each job is compared against every user's AI prompt
- Generates a match score (0-100) and detailed explanation
- Stores results in `user_job_matches` table
- Matches include pros, cons, and key match factors

### Discord Notifications
- When score exceeds user's threshold, notification is queued
- Apprise Lambda sends formatted Discord messages
- Message includes score, job details, pros/cons

### Scheduled Job Fetching
- EventBridge triggers JobSpy Lambda every 30 minutes
- Fetches jobs posted in the last hour
- Scrapes from Indeed, LinkedIn, and Glassdoor

## Database Schema

New columns added to `users`:
- `ai_prompt` - User's ideal job description
- `discord_webhook` - Discord webhook URL for notifications
- `notify_threshold` - Minimum score to trigger notification (default: 70)

New column added to `jobs`:
- `company_info` - JSONB with company research data

New table `user_job_matches`:
- Stores match results between users and jobs
- Includes score, analysis JSON, and notification status

## Infrastructure

| Resource | Terraform File | Purpose |
|----------|---------------|---------|
| EventBridge Rule | `eventbridge.tf` | 30-minute cron trigger |
| Notifier Lambda | `notifier_lambda.tf` | Discord notifications via Apprise |
| SQS Queue | `sqs.tf` | `jobs-to-email` queue for notifications |

## Deployment

To deploy the new features:

```bash
# Apply database migrations
cd backend
go run cmd/migrate/main.go up

# Deploy infrastructure
cd infra/terraform
terraform apply

# Build and deploy Python workers
cd python_workers/notifier
./deploy.sh

cd python_workers/jobspy_fetcher
./deploy.sh
```

## API Endpoints

### New Endpoints

```
GET  /api/users/me           Get user profile with settings
PUT  /api/users/me/prompt    Set AI matching prompt
PUT  /api/users/me/discord   Set Discord webhook URL
PUT  /api/users/me/threshold Set notification threshold
GET  /api/users/me/matches   Get job matches for user
```

## Environment Variables

New variables for Go Lambda:
- `NOTIFY_SQS_QUEUE_URL` - SQS queue URL for notification messages

## Future Improvements

- [ ] Add email notifications (Apprise supports SMTP)
- [ ] Add Slack/Telegram support
- [ ] User dashboard with match history
- [ ] Job deduplication across sources
- [ ] Rate limiting on AI API calls
- [ ] Batch processing for user matching
