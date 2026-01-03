# JobPing Architecture

## Overview

JobPing uses an event-driven architecture with SQS to connect Python Lambda workers with the Go backend for AI-powered job analysis.

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│  Frontend   │      │ API Gateway │      │   Lambdas   │      │  Database   │
│   (React)   │─────▶│   (HTTP)    │─────▶│  (Go/Py)    │─────▶│    (RDS)    │
└─────────────┘      └─────────────┘      └─────────────┘      └─────────────┘
                                                │
                                                ▼
                                          ┌─────────────┐
                                          │  SQS Queue  │
                                          └─────────────┘
```

## Data Flow

1. **User clicks "Fetch Latest Jobs"** 
   - Frontend calls `POST /jobs/fetch`
   - API Gateway routes to JobSpy Lambda (Python)

2. **JobSpy Lambda scrapes jobs**
   - Uses python-jobspy library
   - Fetches from Indeed, LinkedIn, etc.
   - Pushes each job to SQS `jobs-to-filter` queue

3. **Go Lambda processes jobs**
   - Triggered by SQS messages
   - Calls OpenAI API for job analysis
   - Saves results to RDS PostgreSQL

4. **User views results**
   - Frontend calls `GET /api/jobs`
   - Go Lambda reads from RDS
   - Returns AI-analyzed jobs

## Components

### Python Workers (`python_workers/`)

| Worker | Purpose | Trigger |
|--------|---------|---------|
| `jobspy_fetcher` | Scrape jobs from job boards | API Gateway |
| `email_notifier` | Send email notifications (future) | SQS |

### Go Backend (`backend/`)

| Feature | Endpoints | Purpose |
|---------|-----------|---------|
| `user` | `/api/register`, `/api/login`, `/api/preferences` | Auth & user management |
| `job` | `/api/jobs` | Job listing with AI analysis |

### Infrastructure (`infra/terraform/`)

| Resource | File | Purpose |
|----------|------|---------|
| API Gateway | `api_gateway.tf` | HTTP routing |
| Go Lambda | `lambda.tf` | Main API + SQS handler |
| Python Lambda | `python_lambda.tf` | JobSpy fetcher |
| SQS Queues | `sqs.tf` | Message queues |
| RDS | `rds.tf` | PostgreSQL database |

## Local Development

### Prerequisites

- Docker & Docker Compose
- Go 1.25+
- Python 3.11+
- Node.js 18+

### Quick Start

```bash
# Start infrastructure (Postgres + LocalStack)
./scripts/local-dev.sh

# Terminal 1: Start Go backend
cd backend && air

# Terminal 2: Start frontend
cd frontend && npm run dev

# Terminal 3: Test Python worker
cd python_workers/jobspy_fetcher
pip install -r requirements.txt
python test_local.py
```

### URLs

| Service | URL |
|---------|-----|
| Frontend | http://localhost:5173 |
| Go Backend | http://localhost:8080 |
| LocalStack SQS | http://localhost:4566 |
| PostgreSQL | localhost:5433 |

## Deployment

### Deploy Go Lambda

```bash
./scripts/build.sh
./scripts/deploy.sh
```

### Deploy Python Lambda

```bash
./scripts/deploy-python-lambda.sh jobspy_fetcher
```

### Apply Terraform

```bash
cd infra/terraform
terraform plan
terraform apply
```

## Environment Variables

### Go Lambda

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | JWT signing key |
| `OPENAI_API_KEY` | OpenAI API key for job analysis |

### Python Lambda

| Variable | Description |
|----------|-------------|
| `SQS_QUEUE_URL` | SQS queue URL for job messages |
| `AWS_REGION` | AWS region |

## Future Enhancements

- [ ] EventBridge cron trigger (replace manual fetch button)
- [ ] Email notifications via SES
- [ ] User-specific job preferences for AI filtering
- [ ] Multiple job board support (Glassdoor, ZipRecruiter)
- [ ] Job application tracking


