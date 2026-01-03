# EventBridge rule for scheduled job fetching (every 30 minutes)

resource "aws_cloudwatch_event_rule" "job_fetch_cron" {
  name                = "jobping-fetch-jobs"
  description         = "Trigger JobSpy Lambda every 30 minutes to fetch new jobs"
  schedule_expression = "rate(30 minutes)"

  tags = {
    Environment = "production"
    Project     = "jobping"
  }
}

# Target: JobSpy Python Lambda
resource "aws_cloudwatch_event_target" "jobspy_lambda" {
  rule      = aws_cloudwatch_event_rule.job_fetch_cron.name
  target_id = "JobSpyLambda"
  arn       = aws_lambda_function.jobspy_fetcher.arn

  # Default input for scheduled runs (no request body from API Gateway)
  input = jsonencode({
    "source" : "eventbridge",
    "search_term" : "software engineer",
    "location" : "United States",
    "results_wanted" : 10
  })
}

# Permission for EventBridge to invoke JobSpy Lambda
resource "aws_lambda_permission" "eventbridge_invoke_jobspy" {
  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.jobspy_fetcher.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.job_fetch_cron.arn
}

# Output
output "eventbridge_rule_arn" {
  value = aws_cloudwatch_event_rule.job_fetch_cron.arn
}

