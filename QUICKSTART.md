# Quick Start Guide

Get the AI Job Scanner running locally in 5 minutes!

## Prerequisites Check

```bash
# Check Go
go version  # Should be 1.21+

# Check Node
node --version  # Should be 18+

# Check Docker
docker --version
docker-compose --version
```

## 1. Environment Setup

```bash
# Copy environment file
cp .env.example .env

# No changes needed for local development!
```

## 2. Start Database

```bash
# Start PostgreSQL and LocalStack
docker-compose up -d

# Verify they're running
docker-compose ps
```

## 3. Start Backend API

```bash
cd backend

# Download Go dependencies
go mod download

# Start the API server
go run cmd/api/main.go
```

You should see:
```
API server starting on port 8080
```

Keep this terminal open!

## 4. Start Frontend (New Terminal)

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev
```

You should see:
```
VITE ready in XXX ms
Local: http://localhost:5173
```

## 5. Fetch Jobs (New Terminal)

```bash
cd backend

# Run the scanner once
go run cmd/scanner/main.go
```

This will fetch the latest 10 jobs from SpeedyApply API.

## 6. View the App

Open your browser to: **http://localhost:5173**

You should see:
- A beautiful gradient header
- Job cards with company, title, location
- Refresh button to reload jobs

## Testing the API Directly

```bash
# Health check
curl http://localhost:8080/health

# Get jobs
curl http://localhost:8080/api/v1/jobs

# Get specific job
curl http://localhost:8080/api/v1/jobs/{job-id}
```

## Common Issues

### "Database connection failed"
```bash
# Restart Docker services
docker-compose down
docker-compose up -d

# Wait 10 seconds, then try again
```

### "Port 8080 already in use"
```bash
# Find and kill the process
# Windows:
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Mac/Linux:
lsof -ti:8080 | xargs kill -9
```

### "No jobs appearing"
1. Make sure you ran the scanner: `go run cmd/scanner/main.go`
2. Check database: `docker-compose exec postgres psql -U jobscanner -c "SELECT COUNT(*) FROM jobs;"`
3. Check API response: `curl http://localhost:8080/api/v1/jobs`

## What's Running?

- **Frontend**: http://localhost:5173 (React + Vite)
- **Backend API**: http://localhost:8080 (Go)
- **PostgreSQL**: localhost:5432
- **LocalStack**: http://localhost:4566 (Mock AWS)

## Next Steps

1. **Customize the scanner**: Edit `backend/internal/integrations/jobspy/client.go`
2. **Add more features**: See `SETUP.md` for full development guide
3. **Deploy to AWS**: Follow `infra/README.md`

## Development Workflow

```bash
# Terminal 1: API
cd backend && go run cmd/api/main.go

# Terminal 2: Frontend
cd frontend && npm run dev

# Terminal 3: Scanner (optional, runs every 10 min)
cd backend && go run cmd/scanner/main.go
```

## Stopping Everything

```bash
# Stop Docker services
docker-compose down

# Stop API and Frontend: Ctrl+C in their terminals
```

## Need Help?

Check `SETUP.md` for detailed documentation and troubleshooting.
