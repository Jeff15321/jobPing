# JobPing

Job monitoring system that scrapes job postings, enriches them with AI analysis, and notifies users when matching jobs are found.

## Architecture

```
EventBridge (30min) --> JobSpy Lambda --> SQS --> Go Lambda --> PostgreSQL
                                                      |
                                                      v
                                              SQS (notify) --> Apprise Lambda --> Discord
```

**Components:**
- **JobSpy Lambda** (Python): Scrapes jobs from Indeed, LinkedIn via JobSpy library
- **Go Lambda**: AI analysis, user matching, database operations
- **Apprise Lambda** (Python): Sends notifications via Discord webhooks
- **PostgreSQL (RDS)**: Stores jobs, users, and match results

## Quick Start

```bash
# Start database
docker-compose up -d

# Run backend
cd backend && air

# Run frontend
cd frontend && npm install && npm run dev
```

Open http://localhost:5173

## Project Structure

```
jobping/
├── backend/
│   ├── cmd/
│   │   ├── lambda/          # AWS Lambda entrypoint
│   │   └── server/          # Local server entrypoint
│   └── internal/
│       ├── features/
│       │   ├── user/        # Auth, preferences
│       │   └── job/         # Job processing, matching
│       ├── database/        # Migrations, connection
│       └── config/          # Environment config
├── frontend/                # React + TypeScript
├── python_workers/
│   ├── jobspy_fetcher/      # Job scraping
│   └── notifier/            # Discord notifications
└── infra/terraform/         # AWS infrastructure
```

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.21+, Chi router |
| Frontend | React 18, TypeScript, Vite |
| Database | PostgreSQL 16 |
| Infrastructure | AWS Lambda, SQS, RDS, EventBridge |
| IaC | Terraform |

## Deployment

```bash
cd infra/terraform
terraform init
terraform apply
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `OPENAI_API_KEY` | OpenAI API key for job analysis |
| `SQS_QUEUE_URL` | SQS queue for job processing |
| `JWT_SECRET` | Secret for JWT token signing |

## API Endpoints

```
POST /api/auth/register       Register new user
POST /api/auth/login          Login
PUT  /api/users/me/prompt     Set AI matching prompt
PUT  /api/users/me/discord    Set Discord webhook
GET  /api/users/me/matches    Get job matches
GET  /api/jobs                List all jobs
```

## Documentation

- [LOCAL_SETUP.md](LOCAL_SETUP.md) - Detailed local development setup
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - System architecture
- [docs/SQS_EXPLAINED.md](docs/SQS_EXPLAINED.md) - SQS concepts

## License

MIT
