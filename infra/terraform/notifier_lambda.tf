# Notifier Lambda - Python worker for sending Discord notifications via Apprise

# ECR Repository for notifier container
resource "aws_ecr_repository" "notifier" {
  name                 = "jobping-notifier"
  image_tag_mutability = "MUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = false
  }
}

# IAM Role for notifier Lambda
resource "aws_iam_role" "notifier_lambda_role" {
  name = "jobping-notifier-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# CloudWatch Logs policy
resource "aws_iam_role_policy_attachment" "notifier_lambda_logs" {
  role       = aws_iam_role.notifier_lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# SQS access policy for notifier Lambda
resource "aws_iam_role_policy" "notifier_lambda_sqs" {
  name = "jobping-notifier-lambda-sqs-policy"
  role = aws_iam_role.notifier_lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = aws_sqs_queue.jobs_to_email.arn
      }
    ]
  })
}

# Notifier Lambda function
resource "aws_lambda_function" "notifier" {
  function_name = "jobping-notifier"
  role          = aws_iam_role.notifier_lambda_role.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.notifier.repository_url}:latest"
  timeout       = 30
  memory_size   = 256

  environment {
    variables = {
      LOG_LEVEL = "INFO"
    }
  }

  depends_on = [
    aws_iam_role_policy.notifier_lambda_sqs,
    aws_ecr_repository.notifier
  ]

  lifecycle {
    ignore_changes = [image_uri]
  }
}

# SQS trigger for notifier Lambda
resource "aws_lambda_event_source_mapping" "notifier_sqs_trigger" {
  event_source_arn = aws_sqs_queue.jobs_to_email.arn
  function_name    = aws_lambda_function.notifier.arn
  batch_size       = 5
  enabled          = true
}

# Output
output "notifier_lambda_arn" {
  value = aws_lambda_function.notifier.arn
}

output "notifier_ecr_repository_url" {
  value = aws_ecr_repository.notifier.repository_url
}


