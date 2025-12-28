# Running Database Migrations

## Option 1: From Your Local Machine (Easiest - RDS is Public)

Since your RDS is publicly accessible, you can run migrations directly from your local machine:

```bash
# 1. Install migrate tool (if not already installed)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 2. Get database endpoint from Terraform
cd infra/terraform
DB_ENDPOINT=$(terraform output -raw db_endpoint)

# 3. Get password from terraform.tfvars (or use the one you set)
# Read it from terraform.tfvars or use the value you know

# 4. Run migrations
cd ../..  # Back to project root
migrate -path backend/internal/database/migrations \
  -database "postgres://jobscanner:YOUR_PASSWORD@${DB_ENDPOINT}/jobscanner?sslmode=require" \
  up
```

**Replace `YOUR_PASSWORD` with the password from your `terraform.tfvars` file.**

---

## Option 2: From AWS CloudShell

If you want to use CloudShell, you need to upload your code first:

### Step 1: Clone/Upload Your Code to CloudShell

```bash
# Option A: Clone from Git (if your repo is on GitHub/GitLab)
git clone https://github.com/yourusername/jobPing.git
cd jobPing/jobPing

# Option B: Upload via S3 or zip file
# (More complex, not recommended)
```

### Step 2: Install migrate tool

```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

### Step 3: Get Database Endpoint

```bash
# If you have Terraform state in CloudShell
cd infra/terraform
terraform output db_endpoint

# OR use AWS CLI directly
aws rds describe-db-instances \
  --db-instance-identifier jobping-db \
  --query 'DBInstances[0].Endpoint.Address' \
  --output text
```

### Step 4: Run Migrations

```bash
# From project root (after cloning)
migrate -path backend/internal/database/migrations \
  -database "postgres://jobscanner:YOUR_PASSWORD@DB_ENDPOINT:5432/jobscanner?sslmode=require" \
  up
```

---

## Quick Reference

**Get DB endpoint:**
```bash
cd infra/terraform
terraform output db_endpoint
```

**Connection string format:**
```
postgres://jobscanner:PASSWORD@ENDPOINT:5432/jobscanner?sslmode=require
```

**Run migrations:**
```bash
migrate -path backend/internal/database/migrations \
  -database "CONNECTION_STRING" \
  up
```

---

## Troubleshooting

### "connection refused" or timeout
- Check RDS security group allows port 5432 from your IP
- Verify RDS is publicly accessible
- Check RDS status in AWS Console (should be "available")

### "password authentication failed"
- Verify password matches `terraform.tfvars`
- Check username is `jobscanner`

### "database does not exist"
- RDS creates the database automatically, but if it doesn't exist:
  - Connect to RDS and create it: `CREATE DATABASE jobscanner;`

