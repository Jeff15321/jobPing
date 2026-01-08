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

# API Lambda (user endpoints)
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
      ENVIRONMENT  = "production"
      DATABASE_URL = "postgres://jobscanner:${var.db_password}@${aws_db_instance.postgres.endpoint}/jobscanner?sslmode=require"
      JWT_SECRET   = var.jwt_secret
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic,
    aws_db_instance.postgres,
  ]
}

# Jobs API Lambda (job + notification endpoints)
resource "aws_lambda_function" "jobs_api" {
  function_name = "jobping-jobs-api"
  role          = aws_iam_role.lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  timeout       = 30
  memory_size   = 256

  filename         = "${path.module}/../../build/jobs_api.zip"
  source_code_hash = fileexists("${path.module}/../../build/jobs_api.zip") ? filebase64sha256("${path.module}/../../build/jobs_api.zip") : null

  environment {
    variables = {
      ENVIRONMENT  = "production"
      DATABASE_URL = "postgres://jobscanner:${var.db_password}@${aws_db_instance.postgres.endpoint}/jobscanner?sslmode=require"
      JWT_SECRET   = var.jwt_secret
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic,
    aws_db_instance.postgres,
  ]
}

# Stage 1: Job Analysis Worker
resource "aws_lambda_function" "job_analysis_worker" {
  function_name = "jobping-job-analysis-worker"
  role          = aws_iam_role.lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  timeout       = 120
  memory_size   = 512

  filename         = "${path.module}/../../build/job_analysis_worker.zip"
  source_code_hash = fileexists("${path.module}/../../build/job_analysis_worker.zip") ? filebase64sha256("${path.module}/../../build/job_analysis_worker.zip") : null

  environment {
    variables = {
      ENVIRONMENT           = "production"
      DATABASE_URL          = "postgres://jobscanner:${var.db_password}@${aws_db_instance.postgres.endpoint}/jobscanner?sslmode=require"
      OPENAI_API_KEY        = var.openai_api_key
      USER_FANOUT_QUEUE_URL = aws_sqs_queue.user_fanout.url
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic,
    aws_db_instance.postgres,
  ]
}

# Stage 2: User Fanout Worker
resource "aws_lambda_function" "user_fanout_worker" {
  function_name = "jobping-user-fanout-worker"
  role          = aws_iam_role.lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  timeout       = 60
  memory_size   = 256

  filename         = "${path.module}/../../build/user_fanout_worker.zip"
  source_code_hash = fileexists("${path.module}/../../build/user_fanout_worker.zip") ? filebase64sha256("${path.module}/../../build/user_fanout_worker.zip") : null

  environment {
    variables = {
      ENVIRONMENT             = "production"
      DATABASE_URL            = "postgres://jobscanner:${var.db_password}@${aws_db_instance.postgres.endpoint}/jobscanner?sslmode=require"
      USER_ANALYSIS_QUEUE_URL = aws_sqs_queue.user_analysis.url
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic,
    aws_db_instance.postgres,
  ]
}

# Stage 3: User Analysis Worker
resource "aws_lambda_function" "user_analysis_worker" {
  function_name = "jobping-user-analysis-worker"
  role          = aws_iam_role.lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  timeout       = 120
  memory_size   = 512

  filename         = "${path.module}/../../build/user_analysis_worker.zip"
  source_code_hash = fileexists("${path.module}/../../build/user_analysis_worker.zip") ? filebase64sha256("${path.module}/../../build/user_analysis_worker.zip") : null

  environment {
    variables = {
      ENVIRONMENT            = "production"
      DATABASE_URL           = "postgres://jobscanner:${var.db_password}@${aws_db_instance.postgres.endpoint}/jobscanner?sslmode=require"
      OPENAI_API_KEY         = var.openai_api_key
      NOTIFICATION_QUEUE_URL = aws_sqs_queue.notification.url
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic,
    aws_db_instance.postgres,
  ]
}

# Stage 4: Notifier Worker
resource "aws_lambda_function" "notifier_worker" {
  function_name = "jobping-notifier-worker"
  role          = aws_iam_role.lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  timeout       = 30
  memory_size   = 256

  filename         = "${path.module}/../../build/notifier_worker.zip"
  source_code_hash = fileexists("${path.module}/../../build/notifier_worker.zip") ? filebase64sha256("${path.module}/../../build/notifier_worker.zip") : null

  environment {
    variables = {
      ENVIRONMENT  = "production"
      DATABASE_URL = "postgres://jobscanner:${var.db_password}@${aws_db_instance.postgres.endpoint}/jobscanner?sslmode=require"
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

resource "aws_cloudwatch_log_group" "jobs_api" {
  name              = "/aws/lambda/jobping-jobs-api"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "job_analysis_worker" {
  name              = "/aws/lambda/jobping-job-analysis-worker"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "user_fanout_worker" {
  name              = "/aws/lambda/jobping-user-fanout-worker"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "user_analysis_worker" {
  name              = "/aws/lambda/jobping-user-analysis-worker"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "notifier_worker" {
  name              = "/aws/lambda/jobping-notifier-worker"
  retention_in_days = 14
}

output "lambda_function_names" {
  value = {
    api                 = aws_lambda_function.api.function_name
    jobs_api           = aws_lambda_function.jobs_api.function_name
    job_analysis_worker = aws_lambda_function.job_analysis_worker.function_name
    user_fanout_worker  = aws_lambda_function.user_fanout_worker.function_name
    user_analysis_worker = aws_lambda_function.user_analysis_worker.function_name
    notifier_worker     = aws_lambda_function.notifier_worker.function_name
  }
}
