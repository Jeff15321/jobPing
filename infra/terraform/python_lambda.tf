# ECR Repository for Python Lambda images
resource "aws_ecr_repository" "jobspy_fetcher" {
  name                 = "jobping-jobspy-fetcher"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Project = "jobping"
  }
}

# IAM Role for Python Lambda
resource "aws_iam_role" "python_lambda_role" {
  name = "jobping-python-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

# Attach basic Lambda execution policy
resource "aws_iam_role_policy_attachment" "python_lambda_basic" {
  role       = aws_iam_role.python_lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# SQS access policy for Python Lambda
resource "aws_iam_role_policy" "python_lambda_sqs" {
  name = "jobping-python-lambda-sqs-policy"
  role = aws_iam_role.python_lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
          "sqs:GetQueueUrl"
        ]
        Resource = [
          aws_sqs_queue.jobs_to_filter.arn
        ]
      }
    ]
  })
}

# JobSpy Fetcher Lambda Function
resource "aws_lambda_function" "jobspy_fetcher" {
  function_name = "jobping-jobspy-fetcher"
  role          = aws_iam_role.python_lambda_role.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.jobspy_fetcher.repository_url}:latest"
  timeout       = 60
  memory_size   = 512

  environment {
    variables = {
      SQS_QUEUE_URL = aws_sqs_queue.jobs_to_filter.url
      AWS_REGION    = var.aws_region
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.python_lambda_basic,
    aws_iam_role_policy.python_lambda_sqs,
  ]

  # Lifecycle to prevent errors when image doesn't exist yet
  lifecycle {
    ignore_changes = [image_uri]
  }
}

# CloudWatch Logs for Python Lambda
resource "aws_cloudwatch_log_group" "jobspy_fetcher" {
  name              = "/aws/lambda/jobping-jobspy-fetcher"
  retention_in_days = 14
}

# Allow API Gateway to invoke JobSpy Lambda
resource "aws_lambda_permission" "jobspy_api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.jobspy_fetcher.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*/*"
}

output "jobspy_lambda_function_name" {
  value = aws_lambda_function.jobspy_fetcher.function_name
}

output "ecr_repository_url" {
  value = aws_ecr_repository.jobspy_fetcher.repository_url
}

