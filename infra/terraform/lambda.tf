# IAM Role for Lambda
resource "aws_iam_role" "lambda_role" {
  name = "jobping-lambda-role"

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
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Lambda Function
resource "aws_lambda_function" "api" {
  function_name = "jobping-api"
  role          = aws_iam_role.lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  timeout       = 30
  memory_size   = 256

  filename         = "${path.module}/../../build/api.zip"
  source_code_hash = fileexists("${path.module}/../../build/api.zip") ? filebase64sha256("${path.module}/../../build/api.zip") : null

  environment {
    variables = {
      ENVIRONMENT    = "production"
      DATABASE_URL   = "postgres://jobscanner:${var.db_password}@${aws_db_instance.postgres.endpoint}/jobscanner?sslmode=require"
      JWT_SECRET     = var.jwt_secret
      OPENAI_API_KEY = var.openai_api_key
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic,
    aws_db_instance.postgres,
  ]
}

# CloudWatch Logs
resource "aws_cloudwatch_log_group" "api" {
  name              = "/aws/lambda/jobping-api"
  retention_in_days = 14
}

output "lambda_function_name" {
  value = aws_lambda_function.api.function_name
}
