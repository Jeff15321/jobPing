# Deployment Guide - AI Job Scanner
## From Local Development to Production

This guide covers deploying your AI Job Scanner to production environments including AWS Lambda, Vercel, and alternative hosting options.

---

## Table of Contents
1. [Deployment Overview](#deployment-overview)
2. [AWS Deployment (Recommended)](#aws-deployment-recommended)
3. [Vercel Frontend Deployment](#vercel-frontend-deployment)
4. [Alternative Deployment Options](#alternative-deployment-options)
5. [Environment Configuration](#environment-configuration)
6. [Monitoring & Maintenance](#monitoring--maintenance)
7. [Cost Optimization](#cost-optimization)

---

## Deployment Overview

### Architecture Summary
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Vercel        │    │   AWS Lambda    │    │   AWS RDS       │
│   (Frontend)    │───▶│   (Backend)     │───▶│   (Database)    │
│   React App     │    │   Go Services   │    │   PostgreSQL    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   AWS SQS       │
                       │   (Queue)       │
                       └─────────────────┘
```

### Deployment Strategy
- **Frontend**: Vercel (CDN + Edge deployment)
- **Backend**: AWS Lambda (serverless)
- **Database**: AWS RDS PostgreSQL
- **Queue**: AWS SQS
- **Infrastructure**: Terraform (Infrastructure as Code)

---

## AWS Deployment (Recommended)

### Prerequisites

1. **AWS Account** with appropriate permissions
2. **AWS CLI** installed and configured
3. **Terraform** installed (v1.0+)
4. **Go** 1.21+ for building binaries

### Step 1: Configure AWS CLI

```bash
# Install AWS CLI (if not installed)
# Windows: Download from https://aws.amazon.com/cli/
# Mac: brew install awscli
# Linux: sudo apt install awscli

# Configure credentials
aws configure
```

**Required Information:**
- AWS Access Key ID
- AWS Secret Access Key
- Default region (e.g., us-east-1)
- Output format (json)

### Step 2: Deploy Infrastructure with Terraform

```bash
# Navigate to infrastructure directory
cd infra/terraform

# Initialize Terraform
terraform init

# Create terraform.tfvars file
cat > terraform.tfvars << EOF
aws_region  = "us-east-1"
environment = "production"
db_password = "your-secure-password-here"
EOF

# Plan the deployment
terraform plan

# Apply the infrastructure
terraform apply
```

**What gets created:**
- RDS PostgreSQL database
- SQS queue for job processing
- IAM roles for Lambda functions
- Security groups
- VPC subnets (if needed)

**Save the outputs:**
```bash
# Get important values
terraform output db_endpoint
terraform output db_connection_string
terraform output sqs_queue_url
terraform output lambda_role_arn
```

### Step 3: Build and Deploy Lambda Functions

```bash
# Build Go binaries for Lambda
./scripts/build.sh

# This creates:
# - build/api.zip
# - build/scanner.zip  
# - build/matcher.zip
```

### Step 4: Create Lambda Functions

**Option A: Using AWS Console**

1. Go to AWS Lambda Console
2. Create three functions:
   - `jobscanner-api`
   - `jobscanner-scanner`
   - `jobscanner-matcher`

For each function:
- Runtime: `provided.al2`
- Architecture: `x86_64`
- Upload the corresponding zip file
- Set handler to the binary name (e.g., `api`)
- Use the IAM role from Terraform output

**Option B: Using AWS CLI**

```bash
# API Function
aws lambda create-function \
  --function-name jobscanner-api \
  --runtime provided.al2 \
  --role $(terraform output -raw lambda_role_arn) \
  --handler api \
  --zip-file fileb://build/api.zip

# Scanner Function
aws lambda create-function \
  --function-name jobscanner-scanner \
  --runtime provided.al2 \
  --role $(terraform output -raw lambda_role_arn) \
  --handler scanner \
  --zip-file fileb://build/scanner.zip

# Matcher Function
aws lambda create-function \
  --function-name jobscanner-matcher \
  --runtime provided.al2 \
  --role $(terraform output -raw lambda_role_arn) \
  --handler matcher \
  --zip-file fileb://build/matcher.zip
```

### Step 5: Configure Environment Variables

For each Lambda function, set these environment variables:

```bash
# Common variables
ENVIRONMENT=lambda
DATABASE_URL=<from terraform output>
AWS_REGION=us-east-1

# API specific
API_PORT=8080
FRONTEND_URL=https://your-app.vercel.app

# Scanner specific
SPEEDYAPPLY_API_URL=https://api.speedyapply.com
SCAN_INTERVAL_MINUTES=10

# Matcher specific
SQS_QUEUE_URL=<from terraform output>
OPENAI_API_KEY=<your openai key>
EMAIL_FROM=noreply@yourdomain.com
```

### Step 6: Create API Gateway

```bash
# Create REST API
aws apigateway create-rest-api --name jobscanner-api

# Get the API ID from the response
API_ID=<your-api-id>

# Create resources and methods
# This is complex - consider using Terraform or AWS SAM instead
```

**Recommended: Use AWS SAM for API Gateway**

Create `template.yaml`:
```yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31

Resources:
  JobScannerApi:
    Type: AWS::Serverless::Api
    Properties:
      StageName: prod
      Cors:
        AllowMethods: "'GET,POST,PUT,DELETE,OPTIONS'"
        AllowHeaders: "'Content-Type,Authorization'"
        AllowOrigin: "'*'"

  ApiFunction:
    Type: AWS::Serverless::Function
    Properties:
      FunctionName: jobscanner-api
      Events:
        ApiEvent:
          Type: Api
          Properties:
            RestApiId: !Ref JobScannerApi
            Path: /{proxy+}
            Method: ANY
```

Deploy with:
```bash
sam build
sam deploy --guided
```

### Step 7: Set Up Scheduled Scanner

Create EventBridge rule to run scanner every 10 minutes:

```bash
# Create rule
aws events put-rule \
  --name jobscanner-schedule \
  --schedule-expression "rate(10 minutes)"

# Add Lambda target
aws events put-targets \
  --rule jobscanner-schedule \
  --targets "Id"="1","Arn"="arn:aws:lambda:us-east-1:ACCOUNT:function:jobscanner-scanner"

# Add permission
aws lambda add-permission \
  --function-name jobscanner-scanner \
  --statement-id allow-eventbridge \
  --action lambda:InvokeFunction \
  --principal events.amazonaws.com
```

### Step 8: Configure SQS Trigger for Matcher

```bash
# Add SQS trigger to matcher function
aws lambda create-event-source-mapping \
  --event-source-arn $(terraform output -raw sqs_queue_arn) \
  --function-name jobscanner-matcher \
  --batch-size 10
```

---

## Vercel Frontend Deployment

### Step 1: Prepare Frontend

```bash
cd frontend

# Update vercel.json with your API Gateway URL
cat > vercel.json << EOF
{
  "rewrites": [
    {
      "source": "/api/:path*",
      "destination": "https://YOUR-API-GATEWAY-URL.amazonaws.com/prod/api/:path*"
    },
    {
      "source": "/(.*)",
      "destination": "/index.html"
    }
  ]
}
EOF
```

### Step 2: Deploy to Vercel

**Option A: Using Vercel CLI**

```bash
# Install Vercel CLI
npm i -g vercel

# Login to Vercel
vercel login

# Deploy
vercel --prod
```

**Option B: Using GitHub Integration**

1. Push code to GitHub
2. Connect repository to Vercel
3. Configure build settings:
   - Framework: Vite
   - Build Command: `npm run build`
   - Output Directory: `dist`
4. Deploy automatically on push

### Step 3: Configure Environment Variables

In Vercel dashboard, add:
```
VITE_API_URL=https://YOUR-API-GATEWAY-URL.amazonaws.com/prod/api/v1
```

---

## Alternative Deployment Options

### Option 1: Docker + AWS ECS

**Pros:** More control, easier debugging
**Cons:** Higher cost, more management

```dockerfile
# Dockerfile for full stack
FROM node:18-alpine AS frontend
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM golang:1.21-alpine AS backend
WORKDIR /app
COPY backend/go.* ./
RUN go mod download
COPY backend/ ./
RUN go build -o api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=backend /app/api .
COPY --from=frontend /app/dist ./static
EXPOSE 8080
CMD ["./api"]
```

Deploy to ECS:
```bash
# Build and push to ECR
docker build -t jobscanner .
docker tag jobscanner:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/jobscanner:latest
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/jobscanner:latest

# Create ECS service
aws ecs create-service --cluster default --service-name jobscanner --task-definition jobscanner:1 --desired-count 1
```

### Option 2: Traditional VPS (DigitalOcean, Linode)

**Pros:** Cheapest option, full control
**Cons:** Manual management, no auto-scaling

```bash
# On your VPS
sudo apt update
sudo apt install docker.io docker-compose postgresql-client

# Clone repository
git clone https://github.com/yourusername/ai-job-scanner.git
cd ai-job-scanner

# Configure environment
cp .env.example .env
# Edit .env with production values

# Start services
docker-compose -f docker-compose.prod.yml up -d

# Set up reverse proxy (nginx)
sudo apt install nginx
# Configure nginx to proxy to your app
```

### Option 3: Railway/Render (Simplified PaaS)

**Railway:**
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login and deploy
railway login
railway init
railway up
```

**Render:**
1. Connect GitHub repository
2. Configure build settings
3. Add environment variables
4. Deploy automatically

---

## Environment Configuration

### Production Environment Variables

**Backend (.env.production):**
```env
ENVIRONMENT=production
DATABASE_URL=postgres://user:pass@prod-db.amazonaws.com:5432/jobscanner
AWS_REGION=us-east-1
SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789/jobscanner-queue
SPEEDYAPPLY_API_URL=https://api.speedyapply.com
OPENAI_API_KEY=sk-your-production-key
EMAIL_FROM=noreply@yourdomain.com
SES_REGION=us-east-1
FRONTEND_URL=https://your-app.vercel.app
```

**Frontend (.env.production):**
```env
VITE_API_URL=https://api.yourdomain.com/api/v1
```

### Security Considerations

1. **Never commit secrets** to version control
2. **Use AWS Secrets Manager** for sensitive data
3. **Enable HTTPS** everywhere
4. **Configure CORS** properly
5. **Use IAM roles** instead of access keys when possible

---

## Monitoring & Maintenance

### CloudWatch Monitoring

Set up CloudWatch dashboards for:
- Lambda function metrics
- RDS performance
- SQS queue depth
- API Gateway response times

```bash
# Create CloudWatch alarms
aws cloudwatch put-metric-alarm \
  --alarm-name "JobScanner-HighErrorRate" \
  --alarm-description "High error rate in API" \
  --metric-name Errors \
  --namespace AWS/Lambda \
  --statistic Sum \
  --period 300 \
  --threshold 10 \
  --comparison-operator GreaterThanThreshold
```

### Log Management

Configure structured logging:
```go
// In your Go code
log.Printf("Job scan completed: fetched=%d stored=%d", fetched, stored)
```

View logs:
```bash
# Lambda logs
aws logs tail /aws/lambda/jobscanner-api --follow

# RDS logs
aws rds describe-db-log-files --db-instance-identifier jobscanner-db
```

### Backup Strategy

**Database Backups:**
```bash
# Enable automated backups in RDS
aws rds modify-db-instance \
  --db-instance-identifier jobscanner-db \
  --backup-retention-period 7 \
  --apply-immediately
```

**Code Backups:**
- Use Git for version control
- Tag releases: `git tag v1.0.0`
- Store deployment artifacts in S3

---

## Cost Optimization

### AWS Cost Breakdown (Monthly)

| Service | Usage | Cost |
|---------|-------|------|
| RDS db.t3.micro | 24/7 | ~$15 |
| Lambda | 1M requests | ~$2 |
| SQS | 1M messages | ~$0.40 |
| API Gateway | 1M requests | ~$3.50 |
| CloudWatch | Basic monitoring | ~$1 |
| **Total** | | **~$22/month** |

### Optimization Tips

1. **Use Reserved Instances** for RDS (save 30-60%)
2. **Optimize Lambda memory** (test different sizes)
3. **Use SQS batch processing** (reduce Lambda invocations)
4. **Enable RDS auto-scaling** (scale down during low usage)
5. **Use CloudFront** for frontend caching

### Free Tier Benefits

- Lambda: 1M free requests/month
- RDS: 750 hours free (db.t2.micro)
- SQS: 1M free requests/month
- API Gateway: 1M free requests/month

---

## Deployment Checklist

### Pre-Deployment
- [ ] All tests passing locally
- [ ] Environment variables configured
- [ ] Database migrations ready
- [ ] API endpoints tested
- [ ] Frontend builds successfully

### AWS Infrastructure
- [ ] Terraform applied successfully
- [ ] RDS database accessible
- [ ] SQS queue created
- [ ] IAM roles configured
- [ ] Security groups allow necessary traffic

### Lambda Deployment
- [ ] Go binaries built for Linux
- [ ] Lambda functions created
- [ ] Environment variables set
- [ ] API Gateway configured
- [ ] EventBridge schedule created
- [ ] SQS trigger configured

### Frontend Deployment
- [ ] Vercel.json configured with API URL
- [ ] Build successful
- [ ] Environment variables set
- [ ] Domain configured (if custom)

### Post-Deployment
- [ ] Health checks passing
- [ ] API endpoints responding
- [ ] Frontend loading correctly
- [ ] Scanner running on schedule
- [ ] Logs showing expected activity
- [ ] Monitoring alerts configured

### Testing in Production
- [ ] Manual job scan works
- [ ] Jobs display in frontend
- [ ] Database stores jobs correctly
- [ ] Email notifications work (when implemented)

---

## Rollback Strategy

### Quick Rollback
```bash
# Rollback Lambda function
aws lambda update-function-code \
  --function-name jobscanner-api \
  --zip-file fileb://previous-version.zip

# Rollback Vercel deployment
vercel --prod --force
```

### Database Rollback
```bash
# Restore from backup
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier jobscanner-db-restored \
  --db-snapshot-identifier jobscanner-db-snapshot-20231221
```

---

## Troubleshooting Production Issues

### Common Issues

**Lambda Cold Starts:**
- Increase memory allocation
- Use provisioned concurrency for critical functions
- Optimize initialization code

**Database Connection Limits:**
- Use connection pooling
- Monitor active connections
- Consider RDS Proxy

**API Gateway Timeouts:**
- Increase Lambda timeout
- Optimize slow queries
- Add caching layer

**High Costs:**
- Review CloudWatch metrics
- Optimize Lambda memory/timeout
- Use reserved capacity

### Debug Commands

```bash
# Check Lambda function status
aws lambda get-function --function-name jobscanner-api

# View recent logs
aws logs tail /aws/lambda/jobscanner-api --since 1h

# Check RDS status
aws rds describe-db-instances --db-instance-identifier jobscanner-db

# Test API Gateway
curl https://your-api.amazonaws.com/prod/health
```

---

## Next Steps After Deployment

1. **Set up monitoring alerts**
2. **Configure custom domain**
3. **Implement CI/CD pipeline**
4. **Add more comprehensive logging**
5. **Set up staging environment**
6. **Implement feature flags**
7. **Add performance monitoring**

---

**Last Updated:** December 21, 2025
**Tested On:** AWS us-east-1, Vercel, Terraform 1.6+