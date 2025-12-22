# Infrastructure

This directory contains Terraform configuration for AWS infrastructure.

## Prerequisites

- Terraform >= 1.0
- AWS CLI configured with credentials
- AWS account with appropriate permissions

## Setup

1. Initialize Terraform:
```bash
cd infra/terraform
terraform init
```

2. Create a `terraform.tfvars` file:
```hcl
aws_region  = "us-east-1"
environment = "production"
db_password = "your-secure-password-here"
```

3. Plan the infrastructure:
```bash
terraform plan
```

4. Apply the infrastructure:
```bash
terraform apply
```

## Resources Created

- RDS PostgreSQL database
- SQS queue for job processing
- IAM roles for Lambda functions
- Security groups

## Outputs

After applying, you'll get:
- `db_endpoint`: Database endpoint
- `db_connection_string`: Full connection string
- `sqs_queue_url`: SQS queue URL
- `lambda_role_arn`: IAM role ARN for Lambda

## Cost Estimate

- RDS db.t3.micro: ~$15/month
- SQS: ~$0 (free tier covers most usage)
- Lambda: ~$0-5/month (depending on usage)

Total: ~$15-20/month
