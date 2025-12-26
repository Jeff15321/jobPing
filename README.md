# 🤖 AI Job Scanner

An intelligent job monitoring system that fetches jobs from SpeedyApply API, performs AI-powered analysis, and matches them with user preferences using semantic search.

## ✨ Features

- 🔄 **Automated Job Fetching**: Scans SpeedyApply API every 10 minutes
- 🤖 **AI-Powered Analysis**: Web search to gather company reputation, benefits, culture
- 🎯 **Semantic Matching**: Uses AI embeddings to match jobs with user preferences
- 📧 **Email Alerts**: Notifies users when matching jobs are found
- 🎨 **Beautiful UI**: Modern React interface with real-time updates

## 🏗️ Architecture

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│   Scanner   │─────▶│  PostgreSQL  │◀─────│     API     │
│  (Cron/10m) │      │   Database   │      │  (REST)     │
└─────────────┘      └──────────────┘      └─────────────┘
       │                     │                      │
       │                     ▼                      │
       │              ┌─────────────┐              │
       └─────────────▶│  SQS Queue  │              │
                      └─────────────┘              │
                             │                      │
                             ▼                      ▼
                      ┌─────────────┐      ┌─────────────┐
                      │   Matcher   │      │   React UI  │
                      │  (Worker)   │      │  (Vercel)   │
                      └─────────────┘      └─────────────┘
                             │
                             ▼
                      ┌─────────────┐
                      │  Email/SES  │
                      └─────────────┘
```

## 🚀 Quick Start

**Get running in 5 minutes!** See [QUICKSTART.md](QUICKSTART.md)

```bash
# 1. Start services
docker-compose up -d

# 2. Run API (Terminal 1)
cd backend && go run cmd/api/main.go

# 3. Run Frontend (Terminal 2)
cd frontend && npm install && npm run dev

# 4. Fetch jobs (Terminal 3)
cd backend && go run cmd/scanner/main.go

# 5. Open browser
open http://localhost:5173
```

## 🛠️ Tech Stack

### Backend
- **Go 1.21+** - High-performance, concurrent job processing
- **PostgreSQL** - Reliable data storage with JSONB for AI analysis
- **AWS Lambda** - Serverless compute for scalability
- **AWS SQS** - Message queue for job processing
- **AWS SES** - Email notifications

### Frontend
- **React 18** - Modern UI framework
- **TypeScript** - Type-safe development
- **Vite** - Fast build tool
- **Vercel** - Edge deployment

### Infrastructure
- **Terraform** - Infrastructure as Code
- **Docker Compose** - Local development
- **LocalStack** - Local AWS simulation

## 📁 Project Structure

```
ai-job-scanner/
├── backend/
├── frontend/
├── infra/
│   └── terraform/             # AWS infrastructure
├── scripts/
│   ├── build.sh              # Build Go binaries
│   ├── deploy.sh             # Deploy to AWS
│   └── local-dev.sh          # Start local development
└── docker-compose.yml

backend:
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

## 🌐 Deployment

### Local Development
```bash
docker-compose up -d
cd backend && go run cmd/api/main.go
cd frontend && npm run dev
```

### Deploy to AWS
```bash
# Deploy infrastructure
cd infra/terraform
terraform init
terraform apply

# Build and deploy Lambda functions
./scripts/build.sh
./scripts/deploy.sh
```

### Deploy Frontend to Vercel
```bash
cd frontend
vercel deploy --prod
```

## 🔧 Configuration

### Environment Variables

**Local** (`.env`):
```bash
ENVIRONMENT=local
DATABASE_URL=postgres://jobscanner:password@localhost:5432/jobscanner
AWS_ENDPOINT=http://localhost:4566
```

**Production** (AWS Lambda):
```bash
ENVIRONMENT=lambda
DATABASE_URL=<from Terraform>
SQS_QUEUE_URL=<from Terraform>
OPENAI_API_KEY=<your key>
```

## 🧪 Testing

```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm test

# Integration tests
cd tests/integration && go test
```

## 📊 API Endpoints

```
GET  /health                    # Health check
GET  /api/v1/jobs              # List jobs
GET  /api/v1/jobs/:id          # Get job details
POST /api/v1/users             # Create user
PUT  /api/v1/users/:id/preferences  # Update preferences
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test locally
5. Submit a pull request

## 📝 License

See [LICENSE](LICENSE) file for details.

## 🆘 Support

- Check [SETUP.md](SETUP.md) for troubleshooting
- Review [PROJECT_STATUS.md](PROJECT_STATUS.md) for known issues
- Open an issue for bugs or feature requests

## 🎉 Acknowledgments

- SpeedyApply for job data API
- Inspired by the Reddit post on r/csMajors

---

**Ready to get started?** → [QUICKSTART.md](QUICKSTART.md)
