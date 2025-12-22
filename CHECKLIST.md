# Development Checklist

## 🎯 Getting Started

### Prerequisites
- [ ] Go 1.21+ installed (`go version`)
- [ ] Node.js 18+ installed (`node --version`)
- [ ] Docker installed and running (`docker --version`)
- [ ] Git installed (`git --version`)

### Initial Setup
- [ ] Clone the repository
- [ ] Copy `.env.example` to `.env`
- [ ] Review environment variables
- [ ] Read `QUICKSTART.md`

## 🚀 Local Development

### First Time Setup
- [ ] Start Docker services: `docker-compose up -d`
- [ ] Wait 10 seconds for services to initialize
- [ ] Verify PostgreSQL is running: `docker-compose ps`
- [ ] Verify LocalStack is running

### Backend Setup
- [ ] Navigate to backend: `cd backend`
- [ ] Download dependencies: `go mod download`
- [ ] Run go mod tidy: `go mod tidy`
- [ ] Start API server: `go run cmd/api/main.go`
- [ ] Verify API health: `curl http://localhost:8080/health`
- [ ] Test jobs endpoint: `curl http://localhost:8080/api/v1/jobs`

### Frontend Setup
- [ ] Navigate to frontend: `cd frontend`
- [ ] Install dependencies: `npm install`
- [ ] Start dev server: `npm run dev`
- [ ] Open browser: http://localhost:5173
- [ ] Verify UI loads without errors

### Scanner Setup
- [ ] Navigate to backend: `cd backend`
- [ ] Run scanner once: `go run cmd/scanner/main.go`
- [ ] Check for errors in output
- [ ] Verify jobs in database: `docker-compose exec postgres psql -U jobscanner -c "SELECT COUNT(*) FROM jobs;"`
- [ ] Refresh frontend to see jobs

## 🔧 Development Workflow

### Daily Development
- [ ] Start Docker: `docker-compose up -d`
- [ ] Start API: `cd backend && go run cmd/api/main.go`
- [ ] Start Frontend: `cd frontend && npm run dev`
- [ ] Make changes and test

### Testing Changes
- [ ] Backend changes: Restart API server
- [ ] Frontend changes: Auto-reloads via Vite
- [ ] Database changes: Update migrations in `internal/database/database.go`
- [ ] Test API endpoints with curl or Postman
- [ ] Test UI in browser

### Code Quality
- [ ] Format Go code: `cd backend && go fmt ./...`
- [ ] Run Go tests: `cd backend && go test ./...`
- [ ] Lint frontend: `cd frontend && npm run lint`
- [ ] Check TypeScript: `cd frontend && npm run build`

## 📝 Phase 2 Implementation

### AI Integration
- [ ] Choose AI provider (OpenAI/Anthropic)
- [ ] Get API key
- [ ] Implement web search in `internal/integrations/ai/`
- [ ] Test AI analysis on sample job
- [ ] Store results in `jobs.ai_analysis` JSONB field

### Semantic Matching
- [ ] Choose embedding model
- [ ] Implement embedding generation
- [ ] Implement similarity calculation in `internal/domain/match/`
- [ ] Test matching with sample preferences
- [ ] Store matches in `user_job_matches` table

### Email Notifications
- [ ] Setup AWS SES
- [ ] Verify email domain
- [ ] Implement email templates
- [ ] Implement SES client in `internal/integrations/email/`
- [ ] Test email sending locally

### Matcher Service
- [ ] Implement SQS listener in `cmd/matcher/main.go`
- [ ] Connect to job queue
- [ ] Process new jobs
- [ ] Run AI analysis
- [ ] Match with users
- [ ] Send notifications
- [ ] Test end-to-end flow

### User Management
- [ ] Implement user registration API
- [ ] Implement preference update API
- [ ] Create user registration UI
- [ ] Create preference management UI
- [ ] Test user flows

## 🌐 AWS Deployment

### Infrastructure Setup
- [ ] Install Terraform
- [ ] Configure AWS CLI: `aws configure`
- [ ] Review `infra/terraform/main.tf`
- [ ] Create `terraform.tfvars` with your settings
- [ ] Initialize Terraform: `cd infra/terraform && terraform init`
- [ ] Plan infrastructure: `terraform plan`
- [ ] Apply infrastructure: `terraform apply`
- [ ] Save outputs (DB endpoint, SQS URL)

### Lambda Deployment
- [ ] Build binaries: `./scripts/build.sh`
- [ ] Verify zip files in `./build/`
- [ ] Create Lambda functions in AWS Console
- [ ] Upload zip files
- [ ] Configure environment variables
- [ ] Set up EventBridge for scanner (10 min cron)
- [ ] Set up SQS trigger for matcher
- [ ] Test Lambda functions

### API Gateway
- [ ] Create API Gateway
- [ ] Create routes for API endpoints
- [ ] Connect to API Lambda
- [ ] Enable CORS
- [ ] Deploy API
- [ ] Test API Gateway endpoints
- [ ] Save API Gateway URL

### Frontend Deployment
- [ ] Install Vercel CLI: `npm i -g vercel`
- [ ] Update `frontend/vercel.json` with API Gateway URL
- [ ] Set environment variables in Vercel
- [ ] Deploy: `cd frontend && vercel deploy --prod`
- [ ] Test production frontend
- [ ] Verify API integration

## 🧪 Testing

### Unit Tests
- [ ] Write tests for domain logic
- [ ] Write tests for API handlers
- [ ] Write tests for database functions
- [ ] Run all tests: `go test ./...`

### Integration Tests
- [ ] Test API endpoints
- [ ] Test database operations
- [ ] Test job fetching
- [ ] Test matching logic

### End-to-End Tests
- [ ] Test complete job flow
- [ ] Test user registration
- [ ] Test preference updates
- [ ] Test email notifications

## 📊 Monitoring & Optimization

### Monitoring
- [ ] Set up CloudWatch logs
- [ ] Set up CloudWatch metrics
- [ ] Set up error alerts
- [ ] Monitor Lambda costs
- [ ] Monitor RDS costs

### Optimization
- [ ] Optimize database queries
- [ ] Add database indexes
- [ ] Optimize Lambda memory
- [ ] Implement caching
- [ ] Reduce API calls

## 🐛 Troubleshooting

### Common Issues
- [ ] Database connection errors → Check Docker
- [ ] Port conflicts → Kill processes on 8080/5173
- [ ] Go module errors → Run `go mod tidy`
- [ ] Frontend build errors → Delete node_modules, reinstall
- [ ] API CORS errors → Check CORS configuration

### Debug Tools
- [ ] Check Docker logs: `docker-compose logs`
- [ ] Check API logs in terminal
- [ ] Check browser console for frontend errors
- [ ] Use Postman for API testing
- [ ] Use psql for database inspection

## 📚 Documentation

### Keep Updated
- [ ] Update README.md with new features
- [ ] Update API documentation
- [ ] Update deployment guide
- [ ] Add code comments
- [ ] Update environment variable docs

## ✅ Production Ready

### Before Launch
- [ ] All tests passing
- [ ] Error handling implemented
- [ ] Logging configured
- [ ] Monitoring set up
- [ ] Security review completed
- [ ] Performance testing done
- [ ] Documentation complete
- [ ] Backup strategy in place

### Launch
- [ ] Deploy to production
- [ ] Verify all services running
- [ ] Test critical paths
- [ ] Monitor for errors
- [ ] Announce launch!

## 🎉 Post-Launch

### Maintenance
- [ ] Monitor error rates
- [ ] Monitor costs
- [ ] Review user feedback
- [ ] Plan next features
- [ ] Regular security updates

---

**Current Phase**: Phase 1 Complete ✅
**Next Step**: Follow QUICKSTART.md to run locally!
