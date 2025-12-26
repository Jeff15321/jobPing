# Setup Guide - AI Job Scanner

## Quick Start (Local Development)

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL (via Docker)

### Step 1: Clone and Setup

```bash
# Copy environment variables
cp .env.example .env

# Edit .env with your settings (optional for local dev)
```

### Step 2: Start Backend Services

```bash
# Start PostgreSQL and LocalStack
docker-compose up -d postgres localstack

# Wait for services to be ready (about 10 seconds)

# Install Go dependencies
cd backend
go mod download
go mod tidy

# Run database migrations (API will do this automatically)
# Run the API server
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

### Step 3: Start Frontend

```bash
# In a new terminal
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev
```

The frontend will be available at `http://localhost:5173`

### Step 4: Run the Scanner (Optional)

```bash
# In a new terminal
cd backend

# Run the scanner to fetch jobs
go run cmd/scanner/main.go
```

The scanner will fetch jobs every 10 minutes and store them in the database.

## Testing the Setup

1. Open `http://localhost:5173` in your browser
2. You should see the AI Job Scanner interface
3. If no jobs appear, run the scanner manually (Step 4)
4. Check the API health: `curl http://localhost:8080/health`

## Project Structure

```
ai-job-scanner/
├── backend/

├── frontend/
│   ├── src/
│   │   ├── components/        # React components
│   │   ├── services/          # API clients
│   │   └── types/             # TypeScript types
│   └── package.json
├── infra/                     # Terraform for AWS
└── docker-compose.yml

backend:
yourapp/
├── cmd/
│   └── api/
│       └── main.go            # application entrypoint
│
├── internal/
│   ├── features/
│   │   └── jobs/
│   │       ├── http.go        # HTTP handlers
│   │       ├── service.go     # business logic
│   │       ├── repository.go  # DB access
│   │       ├── model.go       # domain entities
│   │       ├── dto.go         # request/response structs
│   │       └── errors.go
│   │
│   ├── server/
│   │   └── routes.go          # route registration
│   │
│   ├── middleware/            # cross-cutting concerns
│   │   ├── auth.go
│   │   ├── logger.go
│   │   └── recovery.go
│   │
│   ├── database/
│   │   ├── db.go              # DB connection setup
│   │   └── migrations/        # future: SQL migrations
│   │
│   ├── config/                # env/config loading
│   │   └── config.go
│   │
│   ├── shared/                # shared utilities (careful!)
│   │   ├── errors.go
│   │   └── pagination.go
│
├── pkg/                       # OPTIONAL: reusable libraries
│
│
├── scripts.go
├── go.mod
├── go.sum
└── README.md

```

## Development Workflow

### Making Changes

1. **Backend changes**: Edit Go files, the server will need to be restarted
2. **Frontend changes**: Vite will hot-reload automatically
3. **Database changes**: Add migrations to `internal/database/database.go`

### Running Tests

```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
npm test
```

## Deployment

### Deploy to AWS

1. **Setup AWS credentials**:
```bash
aws configure
```

2. **Deploy infrastructure**:
```bash
cd infra/terraform
terraform init
terraform apply
```

3. **Build and deploy Lambda functions**:
```bash
./scripts/build.sh
# Then upload the zip files from ./build/ to Lambda
```

### Deploy Frontend to Vercel

1. **Install Vercel CLI**:
```bash
npm i -g vercel
```

2. **Deploy**:
```bash
cd frontend
vercel deploy --prod
```

3. **Update API URL**: Edit `frontend/vercel.json` with your API Gateway URL

## Environment Variables

### Local Development (.env)
```
ENVIRONMENT=local
DATABASE_URL=postgres://jobscanner:password@localhost:5432/jobscanner?sslmode=disable
AWS_ENDPOINT=http://localhost:4566
```

### Production (AWS Lambda)
```
ENVIRONMENT=lambda
DATABASE_URL=<from Terraform output>
SQS_QUEUE_URL=<from Terraform output>
OPENAI_API_KEY=<your key>
```

## Troubleshooting

### Database connection failed
- Ensure PostgreSQL is running: `docker-compose ps`
- Check connection string in `.env`

### Frontend can't reach API
- Verify API is running on port 8080
- Check Vite proxy config in `frontend/vite.config.ts`

### No jobs appearing
- Run the scanner manually: `go run cmd/scanner/main.go`
- Check SpeedyApply API is accessible
- Verify database has jobs: `docker-compose exec postgres psql -U jobscanner -c "SELECT COUNT(*) FROM jobs;"`

## Next Steps

Phase 2 will implement:
- AI web search for job analysis
- Semantic matching with user preferences
- Email notifications via SES
- SQS queue processing

## Support

For issues, check:
- Backend logs: `docker-compose logs api`
- Database: `docker-compose exec postgres psql -U jobscanner`
- Frontend console in browser DevTools
