# JobPing Setup Guide

Complete guide for local testing and AWS deployment of the 4-stage SQS pipeline.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Testing](#local-testing)
   - [Quick Start](#quick-start)
   - [Testing the 4-Stage Pipeline](#testing-the-4-stage-pipeline)
   - [Testing Individual Components](#testing-individual-components)
3. [AWS Deployment](#aws-deployment)
   - [Pre-Deployment Setup](#pre-deployment-setup)
   - [Infrastructure Deployment](#infrastructure-deployment)
   - [Lambda Deployment](#lambda-deployment)
   - [Verification](#verification)

---

## Prerequisites

### Required Tools

| Tool | Version | Install |
|------|---------|---------|
| Docker Desktop | Latest | [docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop) |
| Go | 1.21+ | [go.dev/dl](https://go.dev/dl/) |
| Python | 3.11-3.12 | [python.org](https://www.python.org/downloads/) |
| Node.js | 18+ | [nodejs.org](https://nodejs.org/) |
| Terraform | 1.0+ | [terraform.io/downloads](https://www.terraform.io/downloads) |
| AWS CLI | Latest | [aws.amazon.com/cli](https://aws.amazon.com/cli/) |
| Air (Go hot reload) | Latest | `go install github.com/air-verse/air@latest` |

### Required API Keys

1. **OpenAI API Key** (Optional for local, Required for production)
   - Get from: [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
   - Note: Has usage costs. Mock responses used if not provided.

2. **AWS Account** (For deployment)
   - AWS account with appropriate permissions
   - AWS CLI configured: `aws configure`

---

## Local Testing

### Quick Start

**Option 1: Simple HTTP Testing (No SQS Pipeline)**

This tests the application using HTTP endpoints only (no SQS workers):

```bash
# 1. Start infrastructure
docker-compose up -d

# 2. Set up backend environment
cd backend
cp env.example .env
# Edit .env with your settings (see below)

# 3. Run database migrations
go run ./cmd/server  # This will run migrations on startup

# 4. Start backend (new terminal)
cd backend
air

# 5. Start frontend (new terminal)
cd frontend
npm install
npm run dev

# 6. Test
# Open http://localhost:5173
# Click "Fetch Latest Jobs" (uses mock jobs)
```

**Option 2: Full 4-Stage Pipeline Testing**

This tests the complete SQS pipeline with all workers:

See [Testing the 4-Stage Pipeline](#testing-the-4-stage-pipeline) below.

---

### Environment Setup

**Backend Environment Variables** (`backend/.env`):

```env
# Environment
ENVIRONMENT=local
PORT=8080

# Database (matches docker-compose)
DATABASE_URL=postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable

# JWT (any string for local dev)
JWT_SECRET=my-local-dev-secret-key-12345
JWT_EXPIRY_HOURS=24

# OpenAI (optional - leave empty for mock responses)
OPENAI_API_KEY=sk-your-openai-api-key-here

# SQS Queue URLs (for pipeline testing)
JOB_ANALYSIS_QUEUE_URL=http://localhost:4566/000000000000/jobping-job-analysis
USER_FANOUT_QUEUE_URL=http://localhost:4566/000000000000/jobping-user-fanout
USER_ANALYSIS_QUEUE_URL=http://localhost:4566/000000000000/jobping-user-analysis
NOTIFICATION_QUEUE_URL=http://localhost:4566/000000000000/jobping-notification
```

**Frontend Environment** (optional):

```bash
cd frontend
echo "VITE_API_URL=http://localhost:8080" > .env
```

---

### Testing the 4-Stage Pipeline

This section shows how to test the complete event-driven pipeline locally.

#### Step 1: Start Infrastructure

```bash
# Start PostgreSQL and LocalStack
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

**SQS Queues**: LocalStack automatically creates the 4 queues via `scripts/init-localstack.sh`:
- `jobping-job-analysis`
- `jobping-user-fanout`
- `jobping-user-analysis`
- `jobping-notification`

Verify queues:
```bash
docker exec jobping-localstack awslocal sqs list-queues
```

#### Step 2: Set Up Database

```bash
# Run migrations
cd backend
go run ./cmd/server
# Migrations run automatically on startup
# Press Ctrl+C after migrations complete
```

Or run migrations manually:
```bash
cd backend
go run ./cmd/migrate/main.go up
```

#### Step 3: Create Test Data

**Create a user with AI prompt** (required for pipeline):

```bash
# Register user
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass123"}'

# Login to get token
TOKEN=$(curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass123"}' \
  | jq -r '.token')

# Set AI prompt (required for matching)
curl -X PUT http://localhost:8080/api/user/ai-prompt \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"ai_prompt": "I am looking for a remote software engineering position with Go or Python. I prefer startups with good work-life balance."}'
```

#### Step 4: Start Workers (Terminal Windows)

You need to run each worker in a separate terminal:

**Terminal 1: Job Analysis Worker**
```bash
cd backend
export DATABASE_URL=postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable
export USER_FANOUT_QUEUE_URL=http://localhost:4566/000000000000/jobping-user-fanout
export OPENAI_API_KEY=sk-your-key-here  # Optional
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
go run ./cmd/workers/job_analysis
```

**Terminal 2: User Fanout Worker**
```bash
cd backend
export DATABASE_URL=postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable
export USER_ANALYSIS_QUEUE_URL=http://localhost:4566/000000000000/jobping-user-analysis
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
go run ./cmd/workers/user_fanout
```

**Terminal 3: User Analysis Worker**
```bash
cd backend
export DATABASE_URL=postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable
export NOTIFICATION_QUEUE_URL=http://localhost:4566/000000000000/jobping-notification
export OPENAI_API_KEY=sk-your-key-here  # Optional
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
go run ./cmd/workers/user_analysis
```

**Terminal 4: Notifier Worker**
```bash
cd backend
export DATABASE_URL=postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
go run ./cmd/workers/notifier
```

**Terminal 5: HTTP Server (Optional - for API endpoints)**
```bash
cd backend
air
```

#### Step 5: Test the Pipeline

**Option A: Using Python JobSpy Fetcher**

```bash
cd python_workers/jobspy_fetcher

# Create virtual environment
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Set environment variables
export DATABASE_URL=postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable
export JOB_ANALYSIS_QUEUE_URL=http://localhost:4566/000000000000/jobping-job-analysis
export SQS_ENDPOINT_URL=http://localhost:4566
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test

# Run fetcher (creates jobs in DB and sends to queue)
python -c "
from handler import handler
import json
event = {
    'body': json.dumps({
        'search_term': 'software engineer',
        'location': 'San Francisco, CA',
        'results_wanted': 2
    })
}
handler(event, None)
"
```

**Option B: Manual Testing (Send Job ID to Queue)**

```bash
# 1. Create a job in the database (via HTTP or directly)
# For example, using the mock endpoint:
curl -X POST http://localhost:8080/api/jobs/fetch

# 2. Get a job ID from the database
JOB_ID=$(psql postgres://jobscanner:password@localhost:5433/jobscanner -t -c "SELECT id FROM jobs LIMIT 1" | xargs)

# 3. Send job_id to job-analysis queue
docker exec jobping-localstack awslocal sqs send-message \
  --queue-url http://localhost:4566/000000000000/jobping-job-analysis \
  --message-body "{\"job_id\": \"$JOB_ID\"}"
```

**Watch the Pipeline**:

1. **Stage 1** (Job Analysis): Check Terminal 1 logs - should see company research
2. **Stage 2** (User Fanout): Check Terminal 2 logs - should see users being enqueued
3. **Stage 3** (User Analysis): Check Terminal 3 logs - should see AI matching
4. **Stage 4** (Notification): Check Terminal 4 logs - should see notifications created

**Verify Results**:

```bash
# Check notifications
curl http://localhost:8080/api/notifications

# Check matches in database
psql postgres://jobscanner:password@localhost:5433/jobscanner -c "SELECT * FROM user_job_matches;"

# Check notifications table
psql postgres://jobscanner:password@localhost:5433/jobscanner -c "SELECT * FROM notifications;"
```

---

### Testing Individual Components

#### Test Job Analysis Only

```bash
# Send job_id to job-analysis queue
docker exec jobping-localstack awslocal sqs send-message \
  --queue-url http://localhost:4566/000000000000/jobping-job-analysis \
  --message-body '{"job_id": "your-job-uuid"}'

# Check user-fanout queue for message
docker exec jobping-localstack awslocal sqs receive-message \
  --queue-url http://localhost:4566/000000000000/jobping-user-fanout
```

#### Test User Fanout Only

```bash
# Send job_id to user-fanout queue
docker exec jobping-localstack awslocal sqs send-message \
  --queue-url http://localhost:4566/000000000000/jobping-user-fanout \
  --message-body '{"job_id": "your-job-uuid"}'

# Check user-analysis queue for messages (one per user)
docker exec jobping-localstack awslocal sqs receive-message \
  --queue-url http://localhost:4566/000000000000/jobping-user-analysis
```

#### Test User Analysis Only

```bash
# Send job_id + user_id to user-analysis queue
docker exec jobping-localstack awslocal sqs send-message \
  --queue-url http://localhost:4566/000000000000/jobping-user-analysis \
  --message-body '{"job_id": "job-uuid", "user_id": "user-uuid"}'

# Check notification queue for message (if match found)
docker exec jobping-localstack awslocal sqs receive-message \
  --queue-url http://localhost:4566/000000000000/jobping-notification
```

#### Test Notification Only

```bash
# Send job_id + user_id to notification queue
docker exec jobping-localstack awslocal sqs send-message \
  --queue-url http://localhost:4566/000000000000/jobping-notification \
  --message-body '{"job_id": "job-uuid", "user_id": "user-uuid"}'

# Check notifications table
psql postgres://jobscanner:password@localhost:5433/jobscanner -c "SELECT * FROM notifications;"
```

---

## AWS Deployment

### Pre-Deployment Setup

#### 1. Configure AWS CLI

```bash
aws configure
# Enter your AWS Access Key ID
# Enter your AWS Secret Access Key
# Enter default region (e.g., us-east-1)
# Enter default output format (json)
```

#### 2. Set Up Terraform Variables

Create `infra/terraform/terraform.tfvars`:

```hcl
aws_region      = "us-east-1"
db_password     = "your-secure-database-password"
jwt_secret      = "your-jwt-secret-key"
openai_api_key  = "sk-your-openai-api-key"
```

**Security Note**: Never commit `terraform.tfvars` to git. Add to `.gitignore`.

#### 3. Initialize Terraform

```bash
cd infra/terraform
terraform init
```

---

### Infrastructure Deployment

#### Step 1: Review Terraform Plan

```bash
cd infra/terraform
terraform plan
```

This will show you what resources will be created:
- RDS PostgreSQL database
- 4 SQS queues + 4 DLQs
- 6 Lambda functions
- API Gateway
- IAM roles and policies
- VPC and security groups

#### Step 2: Apply Terraform

```bash
terraform apply
```

Enter `yes` when prompted. This will take 10-15 minutes (mostly RDS creation).

**Output**: Terraform will output important values:
- `api_gateway_url` - Your API endpoint
- `database_endpoint` - RDS database endpoint
- `sqs_queue_urls` - SQS queue URLs

Save these values for later use.

#### Step 3: Run Database Migrations

**Option A: Using Lambda (Recommended)**

Create a temporary migration Lambda or use AWS Systems Manager:

```bash
# Get database endpoint from Terraform output
DB_ENDPOINT=$(terraform output -raw database_endpoint)

# Connect and run migrations
# (You'll need to set up a bastion host or use AWS Systems Manager)
```

**Option B: Using Local Machine**

If your local machine can reach RDS:

```bash
# Get database endpoint and password
DB_ENDPOINT=$(terraform output -raw database_endpoint)
DB_PASSWORD=$(terraform output -raw db_password)

# Set DATABASE_URL
export DATABASE_URL=postgres://jobscanner:${DB_PASSWORD}@${DB_ENDPOINT}:5432/jobscanner?sslmode=require

# Run migrations
cd backend
go run ./cmd/migrate/main.go up
```

**Note**: You may need to add your IP to RDS security group for this to work.

---

### Lambda Deployment

#### Step 1: Build Lambda Binaries

```bash
# From project root
./scripts/build.sh
```

This creates:
- `build/api.zip`
- `build/jobs_api.zip`
- `build/job_analysis_worker.zip`
- `build/user_fanout_worker.zip`
- `build/user_analysis_worker.zip`
- `build/notifier_worker.zip`

#### Step 2: Deploy Go Lambdas

**Note**: The `deploy.sh` script may need updating for the new architecture. Deploy manually:

```bash
# Get Lambda function names from Terraform
cd infra/terraform
terraform output

# Deploy each Lambda individually:

```bash
# Get Lambda function names from Terraform
cd infra/terraform
terraform output lambda_function_names

# Deploy each Lambda
aws lambda update-function-code \
  --function-name jobping-api \
  --zip-file fileb://../build/api.zip

aws lambda update-function-code \
  --function-name jobping-jobs-api \
  --zip-file fileb://../build/jobs_api.zip

aws lambda update-function-code \
  --function-name jobping-job-analysis-worker \
  --zip-file fileb://../build/job_analysis_worker.zip

aws lambda update-function-code \
  --function-name jobping-user-fanout-worker \
  --zip-file fileb://../build/user_fanout_worker.zip

aws lambda update-function-code \
  --function-name jobping-user-analysis-worker \
  --zip-file fileb://../build/user_analysis_worker.zip

aws lambda update-function-code \
  --function-name jobping-notifier-worker \
  --zip-file fileb://../build/notifier_worker.zip
```

#### Step 3: Deploy Python Lambda

**Option A: Using Deployment Script** (if ECR is set up):

```bash
./scripts/deploy-python-lambda.sh jobspy_fetcher
```

**Option B: Manual Deployment** (zip file):

```bash
cd python_workers/jobspy_fetcher
zip -r ../../build/jobspy_fetcher.zip .

aws lambda update-function-code \
  --function-name jobping-jobspy-fetcher \
  --zip-file fileb://../../build/jobspy_fetcher.zip
```

#### Step 4: Update Lambda Environment Variables

Terraform sets most environment variables, but verify:

```bash
# Get queue URLs from Terraform
cd infra/terraform
terraform output sqs_queue_urls

# Update Lambda environment variables if needed
aws lambda update-function-configuration \
  --function-name jobping-job-analysis-worker \
  --environment "Variables={
    DATABASE_URL=postgres://...,
    USER_FANOUT_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/.../jobping-user-fanout,
    OPENAI_API_KEY=sk-...
  }"
```

---

### Verification

#### 1. Test API Endpoints

```bash
# Get API Gateway URL
API_URL=$(cd infra/terraform && terraform output -raw api_gateway_url)

# Test health endpoint
curl ${API_URL}/health

# Test registration
curl -X POST ${API_URL}/api/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass123"}'
```

#### 2. Test SQS Pipeline

```bash
# Get queue URLs
cd infra/terraform
JOB_ANALYSIS_QUEUE=$(terraform output -raw job_analysis_queue_url)

# Send test message
aws sqs send-message \
  --queue-url ${JOB_ANALYSIS_QUEUE} \
  --message-body '{"job_id": "test-job-id"}'
```

#### 3. Check CloudWatch Logs

```bash
# View logs for each Lambda
aws logs tail /aws/lambda/jobping-job-analysis-worker --follow
aws logs tail /aws/lambda/jobping-user-fanout-worker --follow
aws logs tail /aws/lambda/jobping-user-analysis-worker --follow
aws logs tail /aws/lambda/jobping-notifier-worker --follow
```

#### 4. Monitor Dead Letter Queues

```bash
# Check DLQ for failed messages
aws sqs get-queue-attributes \
  --queue-url <dlq-url> \
  --attribute-names ApproximateNumberOfMessages
```

---

## Troubleshooting

### Local Testing

**Issue**: LocalStack queues not created
```bash
# Check init script ran
docker exec jobping-localstack cat /etc/localstack/init/ready.d/init.sh

# Manually create queues
docker exec jobping-localstack awslocal sqs create-queue --queue-name jobping-job-analysis
```

**Issue**: Workers can't connect to database
```bash
# Check database is running
docker-compose ps

# Test connection
docker exec jobping-db psql -U jobscanner -c "SELECT 1"
```

**Issue**: Workers can't send to SQS
```bash
# Check AWS environment variables are set
echo $AWS_ENDPOINT_URL
echo $AWS_ACCESS_KEY_ID

# Test SQS connection
docker exec jobping-localstack awslocal sqs list-queues
```

### AWS Deployment

**Issue**: Terraform fails with permission errors
```bash
# Check AWS credentials
aws sts get-caller-identity

# Verify IAM permissions for Terraform
```

**Issue**: Lambda can't connect to RDS
```bash
# Check security group allows Lambda
# Check VPC configuration
# Verify DATABASE_URL is correct
```

**Issue**: SQS messages not being processed
```bash
# Check Lambda event source mappings
aws lambda list-event-source-mappings

# Check CloudWatch logs for errors
aws logs tail /aws/lambda/jobping-job-analysis-worker
```

---

## Quick Reference

### Local URLs

| Service | URL |
|---------|-----|
| Frontend | http://localhost:5173 |
| Backend API | http://localhost:8080 |
| LocalStack SQS | http://localhost:4566 |
| PostgreSQL | localhost:5433 |

### Local SQS Queue URLs

```
http://localhost:4566/000000000000/jobping-job-analysis
http://localhost:4566/000000000000/jobping-user-fanout
http://localhost:4566/000000000000/jobping-user-analysis
http://localhost:4566/000000000000/jobping-notification
```

### Environment Variables Summary

**Local Development**:
- `DATABASE_URL` - PostgreSQL connection
- `JWT_SECRET` - JWT signing key
- `OPENAI_API_KEY` - OpenAI API key (optional)
- `JOB_ANALYSIS_QUEUE_URL` - SQS queue URL
- `USER_FANOUT_QUEUE_URL` - SQS queue URL
- `USER_ANALYSIS_QUEUE_URL` - SQS queue URL
- `NOTIFICATION_QUEUE_URL` - SQS queue URL
- `AWS_ENDPOINT_URL` - LocalStack endpoint (local only)
- `AWS_ACCESS_KEY_ID` - test (local only)
- `AWS_SECRET_ACCESS_KEY` - test (local only)

**Production**:
- Same as above, but:
  - `AWS_ENDPOINT_URL` - Not set (uses real AWS)
  - `AWS_ACCESS_KEY_ID` - Not needed (uses IAM role)
  - `AWS_SECRET_ACCESS_KEY` - Not needed (uses IAM role)

---

## Next Steps

1. **Set up monitoring**: CloudWatch alarms, SNS notifications
2. **Configure auto-scaling**: Lambda concurrency limits
3. **Set up CI/CD**: GitHub Actions or AWS CodePipeline
4. **Add production notifications**: Email, SMS, webhooks
5. **Optimize costs**: Reserved capacity, SQS batch processing

---

## Support

For issues or questions:
- Check [FEATURES_ISSUES.md](./FEATURES_ISSUES.md) for known issues
- Review [ARCHITECTURE.md](./ARCHITECTURE.md) for system design
- Check individual feature docs in `docs/FEATURES_*.md`

