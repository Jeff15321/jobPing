#!/bin/bash

set -e

echo "Deploying AI Job Scanner to AWS..."

# Build binaries
echo "Step 1: Building binaries..."
./scripts/build.sh

# Deploy infrastructure
echo "Step 2: Deploying infrastructure..."
cd infra/terraform
terraform apply -auto-approve

# Get outputs
DB_ENDPOINT=$(terraform output -raw db_endpoint)
SQS_QUEUE_URL=$(terraform output -raw sqs_queue_url)
LAMBDA_ROLE_ARN=$(terraform output -raw lambda_role_arn)

echo "Infrastructure deployed!"
echo "DB Endpoint: $DB_ENDPOINT"
echo "SQS Queue URL: $SQS_QUEUE_URL"

# TODO: Deploy Lambda functions using AWS CLI or SAM
echo "Step 3: Deploy Lambda functions manually or use AWS SAM"
echo "See README.md for instructions"

cd ../..

echo "Deployment complete!"
echo ""
echo "Next steps:"
echo "1. Update frontend/vercel.json with your API Gateway URL"
echo "2. Deploy frontend: cd frontend && vercel deploy"
echo "3. Configure environment variables in Lambda console"
