#!/bin/bash
# Deploy Python Lambda to AWS

set -e

LAMBDA_NAME=${1:-jobspy_fetcher}
AWS_REGION=${AWS_REGION:-us-east-1}
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ECR_REPO="jobping-${LAMBDA_NAME}"

echo "🚀 Deploying Python Lambda: ${LAMBDA_NAME}"
echo "   AWS Account: ${AWS_ACCOUNT_ID}"
echo "   Region: ${AWS_REGION}"

cd python_workers/${LAMBDA_NAME}

# Build Docker image
echo "📦 Building Docker image..."
docker build -t ${ECR_REPO}:latest .

# Login to ECR
echo "🔐 Logging in to ECR..."
aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

# Tag and push image
echo "📤 Pushing to ECR..."
docker tag ${ECR_REPO}:latest ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPO}:latest
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPO}:latest

# Update Lambda function
echo "🔄 Updating Lambda function..."
aws lambda update-function-code \
  --function-name jobping-${LAMBDA_NAME} \
  --image-uri ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPO}:latest \
  --region ${AWS_REGION}

echo "✅ Deployment complete!"



