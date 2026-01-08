#!/bin/bash
# Initialize LocalStack with SQS queues for 4-stage pipeline

echo "Creating SQS queues for 4-stage pipeline..."

# Stage 1: Job Analysis Queue
awslocal sqs create-queue --queue-name jobping-job-analysis

# Stage 2: User Fanout Queue
awslocal sqs create-queue --queue-name jobping-user-fanout

# Stage 3: User Analysis Queue
awslocal sqs create-queue --queue-name jobping-user-analysis

# Stage 4: Notification Queue
awslocal sqs create-queue --queue-name jobping-notification

echo "SQS queues created successfully!"
awslocal sqs list-queues


