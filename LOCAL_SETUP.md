# Local Development Setup Guide

Complete step-by-step guide to run JobPing locally.

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Docker Desktop | Latest | [docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop) |
| Go | 1.21+ | [go.dev/dl](https://go.dev/dl/) |
| Node.js | 18+ | [nodejs.org](https://nodejs.org/) |
| Python | 3.11-3.12 | [python.org](https://www.python.org/downloads/) (3.13 not supported - NumPy compatibility) |
| Air (Go hot reload) | Latest | `go install github.com/air-verse/air@latest` |

## Step 1: Get API Keys

### OpenAI API Key (Required for AI job analysis)

1. Go to [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
2. Sign up or log in
3. Click "Create new secret key"
4. Copy the key (starts with `sk-...`)

> **Note**: OpenAI API has usage costs. For testing, you can skip this - the app will use mock AI responses.

## Step 2: Configure Environment Variables

### Backend (.env file)

```bash
cd backend
cp env.example .env
```

Edit `backend/.env`:

```env
# Environment
ENVIRONMENT=local
PORT=8080

# Database (matches docker-compose)
DATABASE_URL=postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable

# JWT (can use any string for local dev)
JWT_SECRET=my-local-dev-secret-key-12345
JWT_EXPIRY_HOURS=24

# OpenAI (OPTIONAL - leave empty for mock responses)
OPENAI_API_KEY=sk-your-openai-api-key-here
```

### Frontend (.env file) - Optional

The frontend defaults to `http://localhost:8080`. Only create this if using a different backend URL:

```bash
cd frontend
echo "VITE_API_URL=http://localhost:8080" > .env
```

## Step 3: Start Infrastructure (Docker)

Open a terminal in the project root:

```bash
# Start PostgreSQL and LocalStack (SQS emulator)
docker-compose up -d

# Verify services are running
docker-compose ps
```

Expected output:
```
NAME                 STATUS
jobping-db           running (healthy)
jobping-localstack   running (healthy)
```

### Troubleshooting Docker

```bash
# View logs if services fail
docker-compose logs postgres
docker-compose logs localstack

# Restart services
docker-compose down && docker-compose up -d

# Reset everything (deletes data)
docker-compose down -v && docker-compose up -d
```

## Step 4: Start Go Backend

Open a **new terminal**:

```bash
cd backend

# Install dependencies (first time only)
go mod download

# Run with hot reload
air
```

Expected output:
```
🚀 Server starting on http://localhost:8080
📝 Environment: local
```

### Verify Backend

```bash
curl http://localhost:8080/health
# Should return: {"status":"ok"}
```

## Step 5: Start Frontend

Open a **new terminal**:

```bash
cd frontend

# Install dependencies (first time only)
npm install

# Start dev server
npm run dev
```

Expected output:
```
VITE v5.x.x  ready in xxx ms

➜  Local:   http://localhost:5173/
```

## Step 6: Test the Application

1. Open **http://localhost:5173** in your browser
2. Click **"Fetch Latest Jobs"** button
3. Wait a few seconds for jobs to appear (with AI analysis!)

### What Happens When You Click "Fetch Latest Jobs"

**Local Development:**
```
Frontend → POST /api/jobs/fetch → Go Backend (mock jobs)
                                       ↓
                                 AI Analysis (OpenAI or mock)
                                       ↓
                                 Saves to PostgreSQL
                                       ↓
Frontend ← GET /api/jobs ← Returns AI-analyzed jobs
```

**Production (AWS):**
```
Frontend → POST /jobs/fetch → API Gateway → Python Lambda (JobSpy)
                                                  ↓
                                            SQS Queue
                                                  ↓
                                            Go Lambda (AI analysis)
                                                  ↓
                                            RDS PostgreSQL
```

> **Note**: In local mode, mock job data is used instead of scraping real job boards. This lets you test the full flow without setting up the Python Lambda.
## Step 7: Test Python Worker (Optional)

To test the JobSpy fetcher locally:

```bash
cd python_workers/jobspy_fetcher

# Create virtual environment
python -m venv venv

# Activate (Windows)
venv\Scripts\activate

# Activate (Mac/Linux)
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run test
python test_local.py
## Step 8: Test Full Pipeline with Real Jobs (Optional)

This tests the **complete pipeline** with real job scraping:

```bash
cd python_workers/jobspy_fetcher

# Create virtual environment
python -m venv venv


# Activate (Mac/Linux)
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run full pipeline test (scrapes real jobs → AI analysis → database)
python test_full_pipeline.py
```

This will:
1. **Scrape real jobs** from Indeed using JobSpy
2. **Send each job** to Go backend via HTTP
3. **Run AI analysis** (OpenAI or mock)
4. **Save to PostgreSQL**
5. **View in frontend** at http://localhost:5173

### What Each Test Script Does

| Script | What it tests |
|--------|---------------|
| `test_local.py` | Just JobSpy → LocalStack SQS (no Go backend needed) |
| `test_full_pipeline.py` | **Full flow**: JobSpy → Go backend → AI → PostgreSQL |

## Quick Reference

| Service | URL | Purpose |
|---------|-----|---------|
| Frontend | http://localhost:5173 | React UI |
| Backend API | http://localhost:8080 | Go REST API |
| Health Check | http://localhost:8080/health | API health |
| PostgreSQL | localhost:5433 | Database |
| LocalStack SQS | http://localhost:4566 | SQS emulator |

## Common Issues

### "Failed to connect to database"

```bash
# Check if PostgreSQL is running
docker-compose ps

# Check connection
docker exec -it jobping-db psql -U jobscanner -c "SELECT 1"
```

### "Port 5433 already in use"

```bash
# Find what's using the port
netstat -ano | findstr :5433

# Or change the port in docker-compose.yml
```

### "air: command not found"

```bash
# Install air
go install github.com/air-verse/air@latest

# Add Go bin to PATH (add to your shell profile)
export PATH=$PATH:$(go env GOPATH)/bin
```

### Frontend can't reach backend (CORS error)

The backend already allows CORS from `localhost:5173`. If you're using a different port, update `backend/internal/server/router.go`.

## Environment Variables Reference

### Backend

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ENVIRONMENT` | No | `local` | `local` or `production` |
| `PORT` | No | `8080` | HTTP server port |
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `JWT_SECRET` | Yes | - | Secret for JWT tokens |
| `JWT_EXPIRY_HOURS` | No | `24` | Token expiration |
| `OPENAI_API_KEY` | No | - | OpenAI API key (empty = mock responses) |

### Python Workers (for Lambda deployment)

| Variable | Required | Description |
|----------|----------|-------------|
| `SQS_QUEUE_URL` | Yes | SQS queue URL |
| `SQS_ENDPOINT_URL` | No | LocalStack URL for local testing |
| `AWS_REGION` | No | AWS region (default: us-east-1) |

## Stop Everything

```bash
# Stop Docker services
docker-compose down

# Stop Go backend: Ctrl+C in terminal
# Stop Frontend: Ctrl+C in terminal
```

