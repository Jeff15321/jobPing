# SQS Queue for job processing pipeline
resource "aws_sqs_queue" "jobs_to_filter" {
  name                       = "jobping-jobs-to-filter"
  visibility_timeout_seconds = 60
  message_retention_seconds  = 86400 # 1 day
  receive_wait_time_seconds  = 20    # Long polling

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.jobs_to_filter_dlq.arn
    maxReceiveCount     = 3
  })

  tags = {
    Environment = "production"
    Project     = "jobping"
  }
}

# Dead letter queue for failed messages
resource "aws_sqs_queue" "jobs_to_filter_dlq" {
  name                      = "jobping-jobs-to-filter-dlq"
  message_retention_seconds = 604800 # 7 days

  tags = {
    Environment = "production"
    Project     = "jobping"
  }
}

# SQS Queue for email notifications (placeholder for future)
resource "aws_sqs_queue" "jobs_to_email" {
  name                       = "jobping-jobs-to-email"
  visibility_timeout_seconds = 60
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 20

  tags = {
    Environment = "production"
    Project     = "jobping"
  }
}

# IAM policy for Lambda to access SQS
resource "aws_iam_role_policy" "lambda_sqs" {
  name = "jobping-lambda-sqs-policy"
  role = aws_iam_role.lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:SendMessage"
        ]
        Resource = [
          aws_sqs_queue.jobs_to_filter.arn,
          aws_sqs_queue.jobs_to_email.arn
        ]
      }
    ]
  })
}

# Trigger Go Lambda from SQS
resource "aws_lambda_event_source_mapping" "sqs_to_go_lambda" {
  event_source_arn = aws_sqs_queue.jobs_to_filter.arn
  function_name    = aws_lambda_function.api.arn
  batch_size       = 1
  enabled          = true
}

output "sqs_jobs_to_filter_url" {
  value = aws_sqs_queue.jobs_to_filter.url
}

output "sqs_jobs_to_filter_arn" {
  value = aws_sqs_queue.jobs_to_filter.arn
}

output "sqs_jobs_to_email_url" {
  value = aws_sqs_queue.jobs_to_email.url
}

