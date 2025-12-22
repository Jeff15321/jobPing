# Complete Setup Guide - AI Job Scanner
## From Scratch to Running Server

This guide documents **exactly** what was done to set up the AI Job Scanner from scratch, including all the troubleshooting steps and alternative paths.

---

## Table of Contents
1. [Prerequisites Check](#prerequisites-check)
2. [Initial Setup](#initial-setup)
3. [Docker Setup](#docker-setup)
4. [Backend Setup (Go)](#backend-setup-go)
5. [Frontend Setup (React)](#frontend-setup-react)
6. [Testing the Application](#testing-the-application)
7. [Troubleshooting](#troubleshooting)
8. [Alternative Approaches](#alternative-approaches)

---

## Prerequisites Check

### 1. Check Go Installation
```powershell
go version
```
**Expected Output:** `go version go1.21+`

**If not installed:**
- Download from: https://go.dev/dl/
- Windows: Use the MSI installer
- Mac: Use Homebrew: `brew install go`
- Linux: Use package manager or download binary

### 2. Check Node.js Installation
```powershell
node --version
```
**Expected Output:** `v18.0.0` or higher

**If not installed:**
- Download from: https://nodejs.org/
- Recommended: Use LTS version
- Windows: Use the MSI installer
- Mac: Use Homebrew: `brew install node`
- Linux: Use nvm or package manager

### 3. Check Docker Installation
```powershell
docker --version
docker-compose --version
```
**Expected Output:** 
- `Docker version 20.0+`
- `docker-compose version 1.29+` or `Docker Compose version v2.0+`

**If not installed:**
- Download Docker Desktop: https://www.docker.com/products/docker-desktop
- **IMPORTANT:** Docker Desktop must be running before proceeding

### 4. Verify Docker is Running
```powershell
docker ps
```
**Expected:** Should list running containers (may be empty)

**If error:** Start Docker Desktop and wait 30 seconds for it to initialize

---

## Initial Setup

### 1. Clone or Navigate to Project
```powershell
cd C:\codes\jobPing\jobPing
# Or wherever your project is located
```

### 2. Create Environment File
```powershell
copy .env.example .env
```

### 3. Review Environment Variables
Open `.env` and verify the settings. Default values work for local development:
```env
ENVIRONMENT=local
DATABASE_URL=postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable
API_PORT=8080
FRONTEND_URL=http://localhost:5173
```

**⚠️ IMPORTANT PORT NOTE:**
- We use port **5433** for PostgreSQL (not 5432)
- This avoids conflicts with other PostgreSQL instances (like Supabase)
- If you see "password authentication failed", check the port number

---

## Docker Setup

### 1. Check for Port Conflicts

**Check what's using port 5432:**
```powershell
netstat -an | findstr :5432
```

**If you see multiple entries:**
- Other PostgreSQL instances are running
- This is why we use port 5433 instead

### 2. Start Docker Services

**Option A: Using docker-compose (Recommended)**
```powershell
docker-compose up -d
```

**Option B: If docker-compose fails, use Docker Compose v2:**
```powershell
docker compose up -d
```

**Expected Output:**
```
✔ Network jobping_default          Created
✔ Container jobscanner-db          Started
✔ Container jobscanner-localstack  Started
```

### 3. Verify Services are Running
```powershell
docker-compose ps
```

**Expected Output:**
```
NAME                    STATUS
jobscanner-db           Up (healthy)
jobscanner-localstack   Up
```

### 4. Wait for PostgreSQL to be Ready
PostgreSQL needs 5-10 seconds to initialize. Check health status:
```powershell
docker-compose ps
```
Wait until you see `(healthy)` next to jobscanner-db

### 5. Test Database Connection
```powershell
docker exec jobscanner-db psql -U jobscanner -d jobscanner -c "SELECT 1;"
```

**Expected Output:**
```
 ?column? 
----------
        1
(1 row)
```

**If this fails:**
- Wait another 10 seconds and try again
- Check Docker Desktop to ensure container is running
- Check logs: `docker-compose logs postgres`

---

## Backend Setup (Go)

### 1. Navigate to Backend Directory
```powershell
cd backend
```

### 2. Download Go Dependencies
```powershell
go mod download
```

**If you see errors about go.sum:**
```powershell
# Delete the go.sum file if it exists
Remove-Item go.sum -ErrorAction SilentlyContinue

# Regenerate it
go mod tidy
```

### 3. Copy Environment File to Backend
The Go app looks for `.env` in the backend directory:
```powershell
copy ..\.env .env
```

### 4. Test Database Connection (Optional but Recommended)
Create a test file to verify the connection:
```powershell
go run test_db.go
```

**Expected Output:**
```
Attempting to connect to: postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable
Database opened successfully
Database ping successful!
PostgreSQL version: PostgreSQL 16.11...
```

**If this fails:**
- Check that Docker PostgreSQL is running: `docker-compose ps`
- Verify port 5433 is correct in `.env`
- Check for typos in the connection string

### 5. Start the API Server

**Option A: Run directly (for development)**
```powershell
go run cmd/api/main.go
```

**Expected Output:**
```
2025/12/21 19:47:01 API server starting on port 8080
```

**Option B: Build and run (for production-like testing)**
```powershell
go build -o api.exe cmd/api/main.go
./api.exe
```

### 6. Test the API
Open a **new terminal** and test:
```powershell
# Health check
curl http://localhost:8080/health -UseBasicParsing

# Expected: {"status":"ok"}

# Get jobs (will be empty initially)
curl http://localhost:8080/api/v1/jobs -UseBasicParsing

# Expected: {"count":0,"jobs":null}
```

**If connection refused:**
- Check that the API server is still running in the first terminal
- Verify no other service is using port 8080
- Check firewall settings

---

## Frontend Setup (React)

### 1. Navigate to Frontend Directory
Open a **new terminal** (keep the API server running):
```powershell
cd frontend
```

### 2. Install Dependencies
```powershell
npm install
```

**Expected:** Should install ~200 packages

**If you see warnings:**
- `deprecated` warnings are normal and can be ignored
- `vulnerabilities` warnings are common in development dependencies

**If npm install fails:**
- Try clearing cache: `npm cache clean --force`
- Delete node_modules: `Remove-Item -Recurse -Force node_modules`
- Try again: `npm install`

### 3. Start Development Server
```powershell
npm run dev
```

**Expected Output:**
```
VITE v5.4.21  ready in 382 ms
➜  Local:   http://localhost:5173/
➜  Network: use --host to expose
```

**If port 5173 is in use:**
- Vite will automatically try 5174, 5175, etc.
- Update `FRONTEND_URL` in `.env` if needed

### 4. Open in Browser
Navigate to: **http://localhost:5173**

**Expected:** You should see:
- Beautiful gradient header with "🤖 AI Job Scanner"
- "0 jobs found" message
- "Scan for Jobs" and "Refresh" buttons

---

## Testing the Application

### 1. Test Manual Job Scan

**In the browser:**
1. Click the **"🔍 Scan for Jobs"** button
2. Wait 2-3 seconds
3. You should see: "✅ Scanned successfully! Fetched 5 jobs, stored 5 new jobs."
4. Job cards should appear below

**What's happening:**
- Frontend calls `POST /api/v1/jobs/scan`
- Backend fetches mock jobs (5 sample jobs)
- Jobs are stored in PostgreSQL
- Frontend refreshes and displays them

### 2. Verify Jobs in Database
```powershell
docker exec jobscanner-db psql -U jobscanner -d jobscanner -c "SELECT id, title, company FROM jobs;"
```

**Expected Output:**
```
    id    |          title           | company
----------+--------------------------+---------
 mock-1   | Software Engineer Intern | Google
 mock-2   | Frontend Developer       | Meta
 mock-3   | Backend Engineer         | Netflix
 mock-4   | Full Stack Developer     | Stripe
 mock-5   | DevOps Engineer          | Airbnb
```

### 3. Test API Endpoints

**Get all jobs:**
```powershell
curl http://localhost:8080/api/v1/jobs -UseBasicParsing
```

**Get specific job:**
```powershell
curl http://localhost:8080/api/v1/jobs/mock-1 -UseBasicParsing
```

**Trigger scan via API:**
```powershell
curl -X POST http://localhost:8080/api/v1/jobs/scan -UseBasicParsing
```

---

## Troubleshooting

### Problem: "password authentication failed for user jobscanner"

**Cause:** Wrong port or database not ready

**Solutions:**
1. Check port in `.env` is 5433 (not 5432)
2. Wait 10 seconds for PostgreSQL to initialize
3. Verify Docker container is healthy: `docker-compose ps`
4. Check for port conflicts: `netstat -an | findstr :5433`

**Fix:**
```powershell
# Stop containers
docker-compose down

# Update .env to use port 5433
# DATABASE_URL=postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable

# Restart
docker-compose up -d

# Wait 10 seconds
Start-Sleep -Seconds 10

# Test connection
docker exec jobscanner-db psql -U jobscanner -d jobscanner -c "SELECT 1;"
```

### Problem: "No .env file found"

**Cause:** .env file not in the correct directory

**Solution:**
```powershell
# Copy .env to backend directory
cd backend
copy ..\.env .env
```

### Problem: Docker containers won't start

**Cause:** Docker Desktop not running or port conflicts

**Solutions:**
1. Start Docker Desktop and wait 30 seconds
2. Check for port conflicts:
   ```powershell
   netstat -an | findstr :5433
   netstat -an | findstr :4566
   ```
3. If ports are in use, change them in `docker-compose.yml`

### Problem: Frontend shows "Failed to load jobs"

**Cause:** API server not running or CORS issue

**Solutions:**
1. Verify API is running: `curl http://localhost:8080/health -UseBasicParsing`
2. Check browser console for errors (F12)
3. Verify Vite proxy is configured in `vite.config.ts`

### Problem: "go: module not found"

**Cause:** Go modules not downloaded or corrupted

**Solution:**
```powershell
cd backend
Remove-Item go.sum -ErrorAction SilentlyContinue
go mod tidy
go mod download
```

### Problem: npm install fails

**Solutions:**
```powershell
# Clear npm cache
npm cache clean --force

# Delete node_modules and package-lock.json
Remove-Item -Recurse -Force node_modules
Remove-Item package-lock.json -ErrorAction SilentlyContinue

# Reinstall
npm install
```

---

## Alternative Approaches

### Alternative 1: Run Without Docker

If Docker is causing issues, you can install PostgreSQL directly:

**Install PostgreSQL:**
1. Download from: https://www.postgresql.org/download/
2. Install with default settings
3. Remember the password you set

**Create Database:**
```powershell
# Using psql command line
psql -U postgres
CREATE DATABASE jobscanner;
CREATE USER jobscanner WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE jobscanner TO jobscanner;
\q
```

**Update .env:**
```env
DATABASE_URL=postgres://jobscanner:password@localhost:5432/jobscanner?sslmode=disable
```

### Alternative 2: Use Different Ports

If ports 5433 or 8080 are in use:

**Change PostgreSQL port:**
1. Edit `docker-compose.yml`: Change `"5433:5432"` to `"5434:5432"`
2. Edit `.env`: Change port in `DATABASE_URL` to 5434

**Change API port:**
1. Edit `.env`: Change `API_PORT=8080` to `API_PORT=8081`
2. Edit `frontend/vite.config.ts`: Update proxy target to `http://localhost:8081`

### Alternative 3: Run Everything in Docker

If you prefer to run the API in Docker too:

1. Uncomment the `api` service in `docker-compose.yml`
2. Run: `docker-compose up -d`
3. API will be available at `http://localhost:8080`

**Note:** This requires the Dockerfile to be properly configured.

### Alternative 4: Use Real SpeedyApply API

To use the real API instead of mock data:

1. Get SpeedyApply API endpoint and credentials
2. Update `backend/internal/integrations/jobspy/client.go`
3. Replace the mock data section with real HTTP calls
4. Add API key to `.env` if required

---

## Summary of What's Running

After successful setup, you should have:

| Service | Port | URL | Status Check |
|---------|------|-----|--------------|
| PostgreSQL | 5433 | localhost:5433 | `docker exec jobscanner-db psql -U jobscanner -c "SELECT 1;"` |
| LocalStack | 4566 | localhost:4566 | `curl http://localhost:4566/_localstack/health` |
| Go API | 8080 | http://localhost:8080 | `curl http://localhost:8080/health` |
| React Frontend | 5173 | http://localhost:5173 | Open in browser |

---

## Quick Start Commands (After Initial Setup)

**Start everything:**
```powershell
# Terminal 1: Start Docker services
docker-compose up -d

# Terminal 2: Start API
cd backend
go run cmd/api/main.go

# Terminal 3: Start Frontend
cd frontend
npm run dev
```

**Stop everything:**
```powershell
# Stop Docker
docker-compose down

# Stop API and Frontend: Press Ctrl+C in their terminals
```

---

## Next Steps

1. **Test the scan button** in the browser
2. **View jobs** in the UI
3. **Check database** to see stored jobs
4. **Customize** the mock data in `jobspy/client.go`
5. **Integrate real API** when ready

---

## Common Mistakes to Avoid

1. ❌ **Not starting Docker Desktop** before running docker-compose
2. ❌ **Using port 5432** when another PostgreSQL is running
3. ❌ **Not copying .env to backend directory**
4. ❌ **Not waiting for PostgreSQL to be healthy** before starting API
5. ❌ **Running API from wrong directory** (must be in `backend/`)
6. ❌ **Forgetting to clear environment variables** that override .env

---

## Getting Help

If you're still stuck:

1. Check the logs:
   ```powershell
   docker-compose logs postgres
   docker-compose logs localstack
   ```

2. Verify all services:
   ```powershell
   docker-compose ps
   curl http://localhost:8080/health
   ```

3. Check the detailed error messages in the terminal

4. Review the [TROUBLESHOOTING](#troubleshooting) section above

---

**Last Updated:** December 21, 2025
**Tested On:** Windows 11, Docker Desktop 28.4.0, Go 1.25.5, Node.js 22.13.1
