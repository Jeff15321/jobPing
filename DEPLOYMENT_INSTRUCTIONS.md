# JobPing Deployment Instructions

Complete guide to running locally and deploying to AWS Lambda.

---

## Prerequisites

- **Go** 1.21+
- **Node.js** 18+
- **Docker** (for local PostgreSQL)
- **AWS CLI** (configured with credentials)
- **Terraform** 1.0+

---

## Local Development

### 1. Start PostgreSQL with Docker

```bash
# From project root
docker-compose up -d postgres
```

This starts PostgreSQL on `localhost:5433` with:
- User: `jobscanner`
- Password: `password`
- Database: `jobscanner`

### 2. Start the Backend

```bash
cd backend

# Copy environment file
cp env.example .env

# Run the API (migrations run automatically)
go run cmd/api/main.go
```

The API starts at `http://localhost:8080`

### 3. Start the Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend starts at `http://localhost:5173`

### 4. Test the API

```bash
# Register a new user
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}'

# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}'

# Use the token from login response for protected routes
TOKEN="your-jwt-token-here"

# Create a preference
curl -X POST http://localhost:8080/api/preferences \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"key":"job_title","value":"Software Engineer"}'

# Get all preferences
curl http://localhost:8080/api/preferences \
  -H "Authorization: Bearer $TOKEN"
```

---

## AWS Deployment

### Step 1: Configure AWS CLI

```bash
aws configure
# Enter your AWS Access Key ID, Secret Access Key, and region (e.g., us-east-1)
```

### Step 2: Create terraform.tfvars

```bash
cd infra/terraform

cat > terraform.tfvars << EOF
aws_region   = "us-east-1"
environment  = "production"
db_password  = "your-strong-password-here"
jwt_secret   = "your-super-secret-jwt-key-here"
EOF
```

**Important:** Use a strong, unique password and JWT secret. Never commit this file to git.

### Step 3: Initialize and Apply Terraform (RDS First)

```bash
cd infra/terraform

# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Apply (creates VPC, RDS, IAM roles, etc.)
terraform apply
```

This takes ~10 minutes to create the RDS instance.

**Save the outputs:**
- `db_endpoint` - Your RDS endpoint
- `api_url` - Your API Gateway URL (after Lambda is deployed)

### Step 4: Build the Lambda Function

```bash
# From project root
chmod +x scripts/build.sh
./scripts/build.sh
```

This creates `build/api.zip` containing the Lambda binary.

### Step 5: Deploy Lambda Function

```bash
cd infra/terraform

# Apply again to deploy the Lambda with the new zip
terraform apply
```

### Step 6: Run Database Migrations (One-time)

Option A: Use a bastion host or VPN to access RDS and run migrations manually.

Option B: The Lambda will attempt to run migrations on first request (built into the code).

Option C: Use AWS CloudShell or EC2 in the same VPC:

```bash
# Install migrate CLI
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Run migrations
migrate -path ./backend/internal/database/migrations \
  -database "postgres://jobscanner:YOUR_PASSWORD@YOUR_RDS_ENDPOINT:5432/jobscanner?sslmode=require" \
  up
```

### Step 7: Update Frontend for Production

Create `frontend/.env.production`:

```env
VITE_API_URL=https://your-api-id.execute-api.us-east-1.amazonaws.com
```

Build and deploy frontend:

```bash
cd frontend
npm run build
# Deploy dist/ folder to Vercel, Netlify, S3+CloudFront, etc.
```

### Step 8: Verify Deployment

```bash
# Get your API URL from Terraform output
API_URL=$(terraform output -raw api_url)

# Health check
curl $API_URL/health

# Register a user
curl -X POST $API_URL/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"produser","password":"securepass123"}'
```

---

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check |
| POST | `/api/register` | No | Create new user |
| POST | `/api/login` | No | Login, get JWT token |
| GET | `/api/preferences` | Yes | Get user's preferences |
| POST | `/api/preferences` | Yes | Create a preference |
| PUT | `/api/preferences/{id}` | Yes | Update a preference |
| DELETE | `/api/preferences/{id}` | Yes | Delete a preference |

---

## Environment Variables

### Backend

| Variable | Local Default | Description |
|----------|--------------|-------------|
| `ENVIRONMENT` | `local` | `local` or `production` |
| `PORT` | `8080` | Server port (local only) |
| `DATABASE_URL` | `postgres://...` | PostgreSQL connection string |
| `JWT_SECRET` | `dev-secret...` | JWT signing key |
| `JWT_EXPIRY_HOURS` | `24` | Token expiry in hours |

### Frontend

| Variable | Description |
|----------|-------------|
| `VITE_API_URL` | Backend API URL |

---

## Project Structure

```
backend/
├── cmd/api/main.go           # Application entrypoint
├── internal/
│   ├── config/               # Environment configuration
│   ├── database/             # DB connection & migrations
│   ├── features/users/       # User & preferences feature
│   ├── middleware/           # Auth & logging middleware
│   ├── server/               # Route registration
│   └── shared/               # Shared utilities
├── env.example               # Example environment file
├── go.mod
└── go.sum

frontend/
├── src/
│   ├── App.tsx               # Main app component
│   ├── services/api.ts       # API client
│   └── ...
└── package.json

infra/terraform/
├── main.tf                   # VPC, RDS setup
├── lambda.tf                 # Lambda function & IAM
├── api_gateway.tf            # HTTP API Gateway
└── sqs.tf                    # SQS (future use)
```

---

## Troubleshooting

### Lambda can't connect to RDS

1. Check security groups allow Lambda → RDS on port 5432
2. Verify Lambda is in the same VPC/subnets as RDS
3. Check RDS is publicly accessible or in same subnets

### Migrations fail

1. Verify DATABASE_URL is correct
2. Check if RDS is accessible from where you're running migrations
3. Try connecting with psql first to verify credentials

### CORS errors in browser

1. Check API Gateway CORS configuration
2. Verify frontend is using correct API URL
3. Check Lambda response headers include CORS headers

### Token expired

1. Login again to get a new token
2. Increase `JWT_EXPIRY_HOURS` if needed

---

## Costs Estimate (AWS)

| Service | Estimated Monthly Cost |
|---------|----------------------|
| RDS db.t3.micro | ~$15 |
| Lambda (1M requests) | ~$0.20 |
| API Gateway (1M requests) | ~$1 |
| Data Transfer | ~$1 |
| **Total** | **~$17/month** |

*Costs vary by usage. Use AWS Free Tier for first 12 months.*

---

## Security Checklist

- [ ] Use strong, unique `db_password` and `jwt_secret`
- [ ] Never commit `.env` or `terraform.tfvars` to git
- [ ] Restrict RDS security group to Lambda only (not 0.0.0.0/0)
- [ ] Enable HTTPS only in production
- [ ] Set up AWS CloudWatch alerts for errors
- [ ] Rotate JWT secret periodically

