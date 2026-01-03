# Python Workers

AWS Lambda functions written in Python for the JobPing pipeline.

## Structure

```
python_workers/
├── jobspy_fetcher/       # Fetches jobs from job boards
│   ├── handler.py        # Lambda handler
│   ├── requirements.txt  # Python dependencies
│   ├── Dockerfile        # Container image for Lambda
│   └── test_local.py     # Local testing script
├── email_notifier/       # Email sending (future)
└── shared/               # Shared utilities
    └── sqs_client.py     # SQS client wrapper
```

## Local Development

### Prerequisites

- Python 3.11+
- Docker (for LocalStack)
- pip

### Setup

1. Start LocalStack (from project root):
   ```bash
   docker-compose up -d
   ```

2. Install dependencies:
   ```bash
   cd jobspy_fetcher
   pip install -r requirements.txt
   ```

3. Run local test:
   ```bash
   python test_local.py
   ```

## Deployment

Python Lambdas are deployed as container images:

1. Build the Docker image:
   ```bash
   cd jobspy_fetcher
   docker build -t jobping-jobspy-fetcher .
   ```

2. Push to ECR:
   ```bash
   aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account>.dkr.ecr.us-east-1.amazonaws.com
   docker tag jobping-jobspy-fetcher:latest <account>.dkr.ecr.us-east-1.amazonaws.com/jobping-jobspy-fetcher:latest
   docker push <account>.dkr.ecr.us-east-1.amazonaws.com/jobping-jobspy-fetcher:latest
   ```

3. Update Lambda function:
   ```bash
   aws lambda update-function-code --function-name jobping-jobspy-fetcher --image-uri <account>.dkr.ecr.us-east-1.amazonaws.com/jobping-jobspy-fetcher:latest
   ```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SQS_QUEUE_URL` | URL of the SQS queue to push jobs to |
| `SQS_ENDPOINT_URL` | (Optional) LocalStack endpoint for local testing |
| `AWS_REGION` | AWS region (default: us-east-1) |


