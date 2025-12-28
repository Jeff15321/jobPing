# AWS Lambda Deployment Guide

Complete step-by-step guide to deploy your Go API to AWS Lambda.

---

## Prerequisites Checklist

Before starting, ensure you have:

- [ ] AWS account with billing enabled
- [ ] AWS CLI installed
- [ ] Terraform installed
- [ ] Go 1.21+ installed
- [ ] Git Bash (Windows) or Terminal (Mac/Linux)

---

## Step 1: Install & Configure AWS CLI

### Install AWS CLI

**Windows:**
```bash
# Download from: https://awscli.amazonaws.com/AWSCLIV2.msi
# Or use Chocolatey:
choco install awscli
```

**Mac:**
```bash
brew install awscli
```

**Linux:**
```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
```

### Configure AWS CLI

1. **Get your AWS credentials:**
   - Go to AWS Console → IAM → Users → Your User
   - Click "Security credentials" tab
   - Click "Create access key"
   - Choose "Command Line Interface (CLI)"
   - **SAVE THE ACCESS KEY ID AND SECRET** (you won't see it again!)

2. **Configure AWS CLI:**
```bash
aws configure
```

You'll be prompted for:
```
AWS Access Key ID: [paste your access key]
AWS Secret Access Key: [paste your secret key]
Default region name: us-east-1
Default output format: json
```

3. **Verify configuration:**
```bash
aws sts get-caller-identity
```

You should see your AWS account ID and user ARN.

---

## Step 2: Install Terraform

### Windows
```bash
# Using Chocolatey
choco install terraform

# Or download from: https://developer.hashicorp.com/terraform/downloads
```

### Mac
```bash
brew install terraform
```

### Linux
```bash
wget https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip
unzip terraform_1.6.0_linux_amd64.zip
sudo mv terraform /usr/local/bin/
```

### Verify Installation
```bash
terraform version
```

---

## Step 3: Prepare Terraform Variables

1. **Navigate to terraform directory:**
```bash
cd infra/terraform
```

2. **Create `terraform.tfvars` file:**
```bash
cat > terraform.tfvars << EOF
aws_region   = "us-east-1"
environment  = "production"
database_url  = "YourSuperSecurePassword123!"
jwt_secret   = "your-super-secret-jwt-key-change-this-in-production"
EOF
```

**⚠️ IMPORTANT:**
- Use a **strong password** (mix of uppercase, lowercase, numbers, symbols)
- Use a **strong JWT secret** (at least 32 characters, random)
- **Never commit this file to git!** (it's already in .gitignore)

3. **Generate secure secrets (optional but recommended):**
```bash
# Generate random password
openssl rand -base64 32

# Generate JWT secret
openssl rand -hex 32
```

---

## Step 4: Initialize Terraform

```bash
cd infra/terraform
terraform init
```

This downloads the AWS provider plugin. You should see:
```
Terraform has been successfully initialized!
```

---

## Step 5: Review Terraform Plan

Before applying, review what will be created:

```bash
terraform plan
```

This shows you:
- RDS PostgreSQL instance (db.t3.micro)
- Lambda function
- API Gateway HTTP API
- Security groups
- IAM roles
- CloudWatch log groups

**Expected cost:** ~$15-20/month for RDS + minimal Lambda/API Gateway usage

---

## Step 6: Deploy Infrastructure

```bash
terraform apply
```

Type `yes` when prompted. This takes **10-15 minutes** to create the RDS instance.

**What gets created:**
1. VPC and networking (uses default VPC)
2. RDS PostgreSQL database
3. Security groups
4. Lambda function (empty, will update later)
5. API Gateway HTTP API
6. IAM roles and policies

**Save the outputs:**
```
db_endpoint = "jobscanner-db.xxxxx.us-east-1.rds.amazonaws.com:5432"
api_url = "https://xxxxx.execute-api.us-east-1.amazonaws.com"
```

---

## Step 7: Run Database Migrations

The RDS instance is now running, but it's empty. You need to run migrations.

### Option A: Using AWS CloudShell (Easiest)

1. Go to AWS Console → CloudShell (top right)
2. Install migrate tool:
```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

3. Get your database endpoint:
```bash
cd ~/your-project/infra/terraform
terraform output db_endpoint
```

4. Run migrations:
```bash
migrate -path ~/your-project/backend/internal/database/migrations \
  -database "postgres://jobscanner:YOUR_PASSWORD@YOUR_DB_ENDPOINT:5432/jobscanner?sslmode=require" \
  up
```

### Option B: From Your Local Machine (if RDS is publicly accessible)

```bash
cd backend
# Install migrate if you haven't
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Get DB endpoint from Terraform
cd ../infra/terraform
DB_ENDPOINT=$(terraform output -raw db_endpoint)

# Run migrations
migrate -path ../../backend/internal/database/migrations \
  -database "postgres://jobscanner:YOUR_PASSWORD@${DB_ENDPOINT}/jobscanner?sslmode=require" \
  up
```

---

## Step 8: Build Lambda Function

```bash
# From project root
cd scripts
chmod +x build.sh
./build.sh
```

This creates `build/api.zip` with your Lambda binary.

**Verify the zip was created:**
```bash
ls -lh build/api.zip
# Should be ~5-10MB
```

---

## Step 9: Deploy Lambda Function

### Option A: Using Terraform (Recommended)

The Terraform config already references the zip file. Just update it:

```bash
cd infra/terraform
terraform apply
```

Terraform will detect the new zip file and update the Lambda function.

### Option B: Using AWS CLI

```bash
# Get function name from Terraform
FUNCTION_NAME=$(cd infra/terraform && terraform output -raw lambda_function_name)

# Update Lambda function
aws lambda update-function-code \
  --function-name $FUNCTION_NAME \
  --zip-file fileb://build/api.zip \
  --region us-east-1
```

---

## Step 10: Test Your Deployment

1. **Get your API URL:**
```bash
cd infra/terraform
terraform output api_url
```

2. **Test health endpoint:**
```bash
curl https://YOUR_API_URL/health
# Should return: {"status":"ok"}
```

3. **Test registration:**
```bash
curl -X POST https://YOUR_API_URL/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}'
```

4. **Test login:**
```bash
curl -X POST https://YOUR_API_URL/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}'
```

---

## Step 11: Update Frontend (Optional)

If you have a frontend, update the API URL:

1. **Get API URL:**
```bash
cd infra/terraform
terraform output api_url
```

2. **Update frontend environment:**
```bash
cd ../../frontend
echo "VITE_API_URL=https://YOUR_API_URL" > .env.production
```

3. **Rebuild and deploy frontend:**
```bash
npm run build
# Deploy to Vercel/Netlify/etc.
```

---

## Troubleshooting

### Lambda can't connect to RDS

**Symptoms:** Timeout errors, connection refused

**Fix:**
1. Check security groups allow Lambda → RDS on port 5432
2. Verify Lambda is in same VPC as RDS:
```bash
aws lambda get-function-configuration --function-name jobping-api
# Check VpcConfig is set
```

3. Check RDS security group allows Lambda security group

### "Internal Server Error" in Lambda

**Check CloudWatch Logs:**
```bash
aws logs tail /aws/lambda/jobping-api --follow
```

Common issues:
- Database connection string wrong
- Missing environment variables
- Migration not run

### API Gateway returns 502

**Check:**
1. Lambda function exists and is deployed
2. Lambda has correct handler: `bootstrap`
3. Lambda has correct runtime: `provided.al2023`
4. Check CloudWatch logs for errors

### Database migrations fail

**If RDS is not publicly accessible:**
- Use AWS CloudShell (Option A above)
- Or use a bastion host/EC2 instance in the same VPC

---

## Cost Optimization

**Current setup costs ~$15-20/month:**
- RDS db.t3.micro: ~$15/month
- Lambda: ~$0.20/month (1M requests)
- API Gateway: ~$1/month (1M requests)
- Data transfer: ~$1/month

**To reduce costs:**
- Use RDS Free Tier (if eligible, first 12 months)
- Use smaller RDS instance (db.t3.micro is smallest)
- Monitor CloudWatch for unused resources

---

## Next Steps

1. **Set up monitoring:**
   - CloudWatch alarms for errors
   - CloudWatch dashboards for metrics

2. **Add CI/CD:**
   - GitHub Actions to auto-deploy on push
   - Automated testing before deployment

3. **Security hardening:**
   - Move secrets to AWS Secrets Manager
   - Restrict RDS security group (not 0.0.0.0/0)
   - Enable RDS encryption

4. **Scaling:**
   - Add CloudFront for caching
   - Consider RDS read replicas
   - Add API rate limiting

---

## Quick Reference Commands

```bash
# View infrastructure
cd infra/terraform
terraform show

# Get API URL
terraform output api_url

# Get DB endpoint
terraform output db_endpoint

# View Lambda logs
aws logs tail /aws/lambda/jobping-api --follow

# Update Lambda code
./scripts/build.sh
cd infra/terraform
terraform apply

# Destroy everything (careful!)
terraform destroy
```

---

## Support

If you encounter issues:
1. Check CloudWatch logs
2. Verify all environment variables are set
3. Check security groups allow traffic
4. Ensure migrations have run
5. Verify Lambda function code is deployed

**Remember:** Infrastructure changes take time. Be patient with RDS creation (10-15 minutes)!

