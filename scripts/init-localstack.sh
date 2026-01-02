#!/bin/bash
# Initialize LocalStack with SQS queues

echo "Creating SQS queues..."

# Create jobs-to-filter queue
awslocal sqs create-queue --queue-name jobping-jobs-to-filter

# Create jobs-to-email queue (placeholder for future)
awslocal sqs create-queue --queue-name jobping-jobs-to-email

echo "SQS queues created successfully!"
awslocal sqs list-queues

