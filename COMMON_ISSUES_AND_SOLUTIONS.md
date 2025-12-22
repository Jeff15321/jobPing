# Common Issues and Solutions
## Why Problems Occurred and How to Avoid Them

This document explains the issues we encountered during setup and how to prevent them in the future.

---

## Issue 1: Database Authentication Failed

### What Happened
```
Failed to connect to database: pq: password authentication failed for user "jobscanner"
```

### Root Causes
1. **Port Conflict**: PostgreSQL was trying to use port 5432, but Supabase containers were already using it
2. **Wrong Configuration**: The default config had port 5432, but we changed Docker to use 5433
3. **Environment File Location**: Go app looked for `.env` in `backend/` directory, but we had it in root

### Why This Happens
- **Multiple PostgreSQL instances**: Many developers have Supabase, local PostgreSQL, or other database services running
- **Environment variable precedence**: Go's `godotenv` only loads from current directory
- **Default values**: Code had hardcoded defaults that didn't match our setup

### How to Avoid
```powershell
# 1. Always check for port conflicts BEFORE starting
netstat -an | findstr :5432

# 2. Use non-standard ports for development
# In docker-compose.yml: "5433:5432" instead of "5432:5432"

# 3. Always copy .env to backend directory
copy .env backend\.env

# 4. Verify database connection before starting API
docker exec jobscanner-db psql -U jobscanner -d jobscanner -c "SELECT 1;"
```

### Prevention Checklist
- [ ] Check port availability: `netstat -an | findstr :5432`
- [ ] Use unique ports (5433, 5434, etc.)
- [ ] Copy .env to all directories that need it
- [ ] Test database connection before starting API

---

## Issue 2: Go Module and Import Errors

### What Happened
```
malformed go.sum: wrong number of fields
"encoding/json" imported and not used
undefined: http
```

### Root Causes
1. **Corrupted go.sum**: The initial go.sum file had placeholder text instead of actual checksums
2. **Unused imports**: When switching from real API to mock data, some imports became unused
3. **Missing imports**: Removed too many imports, breaking the code

### Why This Happens
- **Copy-paste errors**: When creating files manually, easy to make mistakes
- **Refactoring**: Changing code without updating imports
- **Go's strict import rules**: Go compiler rejects unused imports

### How to Avoid
```powershell
# 1. Always regenerate go.sum when there are issues
Remove-Item go.sum -ErrorAction SilentlyContinue
go mod tidy
go mod download

# 2. Use IDE with Go support (VS Code with Go extension)
# It will automatically manage imports

# 3. Test compilation after every change
go build ./...

# 4. Use go fmt to format code properly
go fmt ./...
```

### Prevention Checklist
- [ ] Use IDE with Go language support
- [ ] Run `go mod tidy` after adding dependencies
- [ ] Test compilation frequently: `go build ./...`
- [ ] Use `go fmt` to format code

---

## Issue 3: Port Already in Use

### What Happened
```
listen tcp :8080: bind: Only one usage of each socket address normally permitted
```

### Root Causes
1. **Previous API instance still running**: Didn't properly stop the old server
2. **Background processes**: API was running in background from previous attempts
3. **No process management**: Manually started processes without tracking them

### Why This Happens
- **Development workflow**: Restarting servers frequently during development
- **Terminal management**: Multiple terminals, easy to lose track
- **Windows process behavior**: Processes don't always terminate cleanly

### How to Avoid
```powershell
# 1. Always check what's using a port before starting
netstat -ano | findstr :8080

# 2. Kill processes properly
taskkill /PID <PID> /F

# 3. Use process management tools
# - Use controlPwshProcess for background services
# - Use task manager to monitor processes

# 4. Standardize shutdown procedure
# Always Ctrl+C to stop servers gracefully
```

### Prevention Checklist
- [ ] Check port availability: `netstat -ano | findstr :8080`
- [ ] Kill old processes before starting new ones
- [ ] Use Ctrl+C to stop servers gracefully
- [ ] Keep track of running processes

---

## Issue 4: Docker Services Not Ready

### What Happened
- Database connection failed even though container was running
- Services started but weren't accepting connections

### Root Causes
1. **Timing issue**: PostgreSQL needs time to initialize after container starts
2. **Health check ignored**: Didn't wait for health check to pass
3. **Impatience**: Tried to connect immediately after `docker-compose up`

### Why This Happens
- **Container vs Service ready**: Container can be "running" but service not ready
- **Database initialization**: PostgreSQL needs to create databases, users, etc.
- **Network setup**: Docker networking takes time to establish

### How to Avoid
```powershell
# 1. Always wait for health checks
docker-compose up -d
Start-Sleep -Seconds 10
docker-compose ps  # Check for "(healthy)" status

# 2. Test service availability
docker exec jobscanner-db psql -U jobscanner -d jobscanner -c "SELECT 1;"

# 3. Use health checks in docker-compose.yml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U jobscanner"]
  interval: 5s
  timeout: 5s
  retries: 5
```

### Prevention Checklist
- [ ] Wait 10 seconds after `docker-compose up -d`
- [ ] Check health status: `docker-compose ps`
- [ ] Test service connectivity before proceeding
- [ ] Use proper health checks in Docker configuration

---

## Issue 5: Frontend API Connection Failed

### What Happened
```
Error: Failed to scan jobs: Not Found
```

### Root Causes
1. **API server not restarted**: New endpoints weren't available in old server
2. **Code changes not applied**: Modified handlers but didn't restart server
3. **Endpoint mismatch**: Frontend calling endpoint that doesn't exist

### Why This Happens
- **Development workflow**: Easy to forget to restart server after code changes
- **Go compilation**: Unlike interpreted languages, Go needs restart for changes
- **API versioning**: Endpoints change during development

### How to Avoid
```powershell
# 1. Always restart API server after code changes
# Stop with Ctrl+C, then restart with:
go run cmd/api/main.go

# 2. Test endpoints after changes
curl http://localhost:8080/api/v1/jobs/scan -Method POST

# 3. Use hot reload tools for development
# Consider using air: https://github.com/cosmtrek/air

# 4. Verify endpoint exists
curl http://localhost:8080/api/v1/ -UseBasicParsing
```

### Prevention Checklist
- [ ] Restart API server after code changes
- [ ] Test new endpoints with curl
- [ ] Check browser console for detailed errors
- [ ] Verify API routes are registered correctly

---

## Issue 6: Environment Variables Not Loading

### What Happened
- `.env` file existed but values weren't being used
- Default values used instead of configured values

### Root Causes
1. **File location**: `.env` in wrong directory
2. **File format**: Incorrect syntax in `.env` file
3. **Caching**: Environment variables cached from previous runs

### Why This Happens
- **Working directory**: Go apps load `.env` from current working directory
- **Syntax errors**: Spaces around `=`, quotes, etc.
- **OS environment**: System environment variables override `.env`

### How to Avoid
```powershell
# 1. Always copy .env to the directory where you run the app
copy .env backend\.env

# 2. Verify .env syntax (no spaces around =)
# Good: DATABASE_URL=postgres://...
# Bad:  DATABASE_URL = postgres://...

# 3. Clear environment variables if needed
$env:DATABASE_URL = $null

# 4. Test configuration loading
go run debug_config.go  # Create a test script
```

### Prevention Checklist
- [ ] Copy `.env` to working directory
- [ ] Check `.env` syntax (no spaces around `=`)
- [ ] Clear conflicting environment variables
- [ ] Test configuration loading with debug script

---

## General Prevention Strategies

### 1. Use Consistent Development Environment
```powershell
# Create a startup script
# startup.ps1
docker-compose up -d
Start-Sleep -Seconds 10
cd backend
copy ..\.env .env
go run cmd/api/main.go
```

### 2. Implement Health Checks
```powershell
# Create a health check script
# health-check.ps1
Write-Host "Checking Docker services..."
docker-compose ps

Write-Host "Checking API..."
curl http://localhost:8080/health -UseBasicParsing

Write-Host "Checking Frontend..."
curl http://localhost:5173 -UseBasicParsing
```

### 3. Use Process Management
```powershell
# Use background processes for services
controlPwshProcess -action start -command "go run cmd/api/main.go" -path backend
controlPwshProcess -action start -command "npm run dev" -path frontend
```

### 4. Document Your Setup
- Keep a personal setup checklist
- Document any custom configurations
- Note which ports you're using
- Track running processes

### 5. Use Development Tools
- **VS Code with Go extension**: Automatic import management
- **Docker Desktop**: Visual container management
- **Postman**: API testing
- **Browser DevTools**: Frontend debugging

---

## Quick Reference: Common Commands

### Check What's Running
```powershell
# Check ports
netstat -ano | findstr :8080
netstat -ano | findstr :5433

# Check Docker
docker-compose ps
docker ps

# Check processes
Get-Process | Where-Object {$_.ProcessName -like "*go*"}
```

### Clean Restart Everything
```powershell
# Stop everything
docker-compose down
taskkill /IM go.exe /F  # Kill any Go processes

# Clean start
docker-compose up -d
Start-Sleep -Seconds 10
cd backend
copy ..\.env .env
go run cmd/api/main.go
```

### Emergency Reset
```powershell
# Nuclear option: reset everything
docker-compose down -v  # Remove volumes too
docker system prune -f  # Clean Docker
Remove-Item backend\.env -ErrorAction SilentlyContinue
Remove-Item backend\go.sum -ErrorAction SilentlyContinue
go mod tidy
```

---

## Key Takeaways

1. **Always check prerequisites** before starting
2. **Use unique ports** to avoid conflicts
3. **Wait for services** to be ready before connecting
4. **Restart servers** after code changes
5. **Copy configuration files** to correct locations
6. **Test each step** before proceeding to the next
7. **Keep track of running processes**
8. **Use proper shutdown procedures**

**Remember**: Most issues are caused by timing, configuration, or process management. Following a consistent workflow prevents 90% of problems.

---

**Pro Tip**: Create a personal checklist based on this document and follow it every time you start development. It takes 2 minutes but saves hours of debugging!