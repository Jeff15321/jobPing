# Quick Start Guide - AI Job Scanner
## Get Running in 5 Minutes

This is a summary of the exact steps to get the AI Job Scanner running from scratch.

---

## Prerequisites (2 minutes)

**Check you have these installed:**
```powershell
go version        # Should be 1.21+
node --version    # Should be 18+
docker --version  # Should be 20+
```

**Start Docker Desktop** and wait 30 seconds for it to initialize.

---

## Setup (3 minutes)

### 1. Environment Setup
```powershell
# Navigate to project
cd C:\codes\jobPing\jobPing

# Copy environment file
copy .env.example .env

# Copy to backend directory (important!)
copy .env backend\.env
```

### 2. Start Database Services
```powershell
# Start PostgreSQL and LocalStack
docker-compose up -d

# Wait 10 seconds for PostgreSQL to initialize
Start-Sleep -Seconds 10

# Verify services are running
docker-compose ps
```

**Expected:** You should see `jobscanner-db` with status `Up (healthy)`

### 3. Start Backend API
```powershell
# Navigate to backend
cd backend

# Download Go dependencies
go mod download
go mod tidy

# Start API server
go run cmd/api/main.go
```

**Expected:** `2025/12/21 19:56:27 API server starting on port 8080`

**Keep this terminal open!**

### 4. Start Frontend (New Terminal)
```powershell
# Navigate to frontend
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

**Expected:** `VITE ready in XXX ms ➜ Local: http://localhost:5173/`

---

## Test It Works

### 1. Open Browser
Go to: **http://localhost:5173**

You should see:
- Beautiful gradient header "🤖 AI Job Scanner"
- "0 jobs found" message
- Green "🔍 Scan for Jobs" button

### 2. Test Job Scanning
1. Click **"🔍 Scan for Jobs"**
2. Wait 2-3 seconds
3. You should see: "✅ Scanned successfully! Fetched 5 jobs, stored 5 new jobs."
4. 5 job cards appear (Google, Meta, Netflix, Stripe, Airbnb)

### 3. Verify API (Optional)
```powershell
# Health check
curl http://localhost:8080/health -UseBasicParsing

# Get jobs
curl http://localhost:8080/api/v1/jobs -UseBasicParsing
```

---

## What's Running

| Service | Port | URL | Purpose |
|---------|------|-----|---------|
| PostgreSQL | 5433 | localhost:5433 | Database |
| LocalStack | 4566 | localhost:4566 | Mock AWS |
| Go API | 8080 | http://localhost:8080 | Backend API |
| React Frontend | 5173 | http://localhost:5173 | User Interface |

---

## Daily Workflow

**To start everything:**
```powershell
# Terminal 1: Start Docker (if not running)
docker-compose up -d

# Terminal 2: Start API
cd backend
go run cmd/api/main.go

# Terminal 3: Start Frontend
cd frontend
npm run dev
```

**To stop everything:**
```powershell
# Stop Docker
docker-compose down

# Stop API and Frontend: Press Ctrl+C in their terminals
```

---

## Quick Troubleshooting

**Problem:** "password authentication failed"
**Fix:** Check that PostgreSQL is healthy: `docker-compose ps`

**Problem:** "port 8080 already in use"
**Fix:** Kill existing process: `taskkill /PID <PID> /F`

**Problem:** "No .env file found"
**Fix:** Copy .env to backend: `copy .env backend\.env`

**Problem:** Frontend shows "Failed to load jobs"
**Fix:** Verify API is running: `curl http://localhost:8080/health`

---

## Success Checklist

- [ ] Docker containers are healthy: `docker-compose ps`
- [ ] API responds: `curl http://localhost:8080/health`
- [ ] Frontend loads: http://localhost:5173
- [ ] Scan button works and shows 5 jobs
- [ ] Jobs are stored in database

**If all checkboxes are ✅, you're ready to develop!**

---

## File Structure Overview

```
ai-job-scanner/
├── backend/           # Go API server
│   ├── cmd/api/       # Main API application
│   ├── internal/      # Business logic
│   └── .env          # Backend environment (copy from root)
├── frontend/          # React application
│   ├── src/          # React components
│   └── package.json  # Dependencies
├── docker-compose.yml # Database services
├── .env              # Main environment file
└── README.md         # Project documentation
```

---

**Total Time:** ~5 minutes
**Next Step:** Read `DEPLOYMENT_GUIDE.md` to deploy to production