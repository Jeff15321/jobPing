# Stage 1: Job Analysis Queue
resource "aws_sqs_queue" "job_analysis" {
  name                       = "jobping-job-analysis"
  visibility_timeout_seconds = 120  # 2 min for AI calls
  message_retention_seconds  = 86400 # 1 day
  receive_wait_time_seconds  = 20    # Long polling

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.job_analysis_dlq.arn
    maxReceiveCount     = 3
  })

  tags = {
    Environment = "production"
    Project     = "jobping"
  }
}

resource "aws_sqs_queue" "job_analysis_dlq" {
  name                      = "jobping-job-analysis-dlq"
  message_retention_seconds = 604800 # 7 days

  tags = {
    Environment = "production"
    Project     = "jobping"
  }
}

# Stage 2: User Fanout Queue
resource "aws_sqs_queue" "user_fanout" {
  name                       = "jobping-user-fanout"
  visibility_timeout_seconds = 60
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 20

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.user_fanout_dlq.arn
    maxReceiveCount     = 3
  })

  tags = {
    Environment = "production"
    Project     = "jobping"
  }
}

resource "aws_sqs_queue" "user_fanout_dlq" {
  name                      = "jobping-user-fanout-dlq"
  message_retention_seconds = 604800

  tags = {
    Environment = "production"
    Project     = "jobping"
  }
}

# Stage 3: User Analysis Queue
resource "aws_sqs_queue" "user_analysis" {
  name                       = "jobping-user-analysis"
  visibility_timeout_seconds = 120  # 2 min for AI calls
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 20

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.user_analysis_dlq.arn
    maxReceiveCount     = 3
  })

  tags = {
    Environment = "production"
    Project     = "jobping"
  }
}

resource "aws_sqs_queue" "user_analysis_dlq" {
  name                      = "jobping-user-analysis-dlq"
  message_retention_seconds = 604800

  tags = {
    Environment = "production"
    Project     = "jobping"
  }
}

# Stage 4: Notification Queue
resource "aws_sqs_queue" "notification" {
  name                       = "jobping-notification"
  visibility_timeout_seconds = 30
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 20

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.notification_dlq.arn
    maxReceiveCount     = 3
  })

  tags = {
    Environment = "production"
    Project     = "jobping"
  }
}

resource "aws_sqs_queue" "notification_dlq" {
  name                      = "jobping-notification-dlq"
  message_retention_seconds = 604800

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
          aws_sqs_queue.job_analysis.arn,
          aws_sqs_queue.user_fanout.arn,
          aws_sqs_queue.user_analysis.arn,
          aws_sqs_queue.notification.arn
        ]
      }
    ]
  })
}

# SQS Event Source Mappings
resource "aws_lambda_event_source_mapping" "job_analysis" {
  event_source_arn = aws_sqs_queue.job_analysis.arn
  function_name    = aws_lambda_function.job_analysis_worker.arn
  batch_size       = 1
  enabled          = true
}

resource "aws_lambda_event_source_mapping" "user_fanout" {
  event_source_arn = aws_sqs_queue.user_fanout.arn
  function_name    = aws_lambda_function.user_fanout_worker.arn
  batch_size       = 1
  enabled          = true
}

resource "aws_lambda_event_source_mapping" "user_analysis" {
  event_source_arn = aws_sqs_queue.user_analysis.arn
  function_name    = aws_lambda_function.user_analysis_worker.arn
  batch_size       = 1
  enabled          = true
}

resource "aws_lambda_event_source_mapping" "notification" {
  event_source_arn = aws_sqs_queue.notification.arn
  function_name    = aws_lambda_function.notifier_worker.arn
  batch_size       = 1
  enabled          = true
}

# Outputs
output "sqs_job_analysis_url" {
  value = aws_sqs_queue.job_analysis.url
}

output "sqs_user_fanout_url" {
  value = aws_sqs_queue.user_fanout.url
}

output "sqs_user_analysis_url" {
  value = aws_sqs_queue.user_analysis.url
}

output "sqs_notification_url" {
  value = aws_sqs_queue.notification.url
}
