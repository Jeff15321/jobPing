# HTTP API Gateway (simple setup)
resource "aws_apigatewayv2_api" "api" {
  name          = "jobping-api"
  protocol_type = "HTTP"

  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers = ["Content-Type", "Authorization"]
    max_age       = 300
  }
}

# Connect API Gateway to User API Lambda
resource "aws_apigatewayv2_integration" "api" {
  api_id             = aws_apigatewayv2_api.api.id
  integration_type   = "AWS_PROXY"
  integration_uri    = aws_lambda_function.api.invoke_arn
  integration_method = "POST"
}

# Connect API Gateway to Jobs API Lambda
resource "aws_apigatewayv2_integration" "jobs_api" {
  api_id             = aws_apigatewayv2_api.api.id
  integration_type   = "AWS_PROXY"
  integration_uri    = aws_lambda_function.jobs_api.invoke_arn
  integration_method = "POST"
}

# Connect API Gateway to JobSpy Lambda
resource "aws_apigatewayv2_integration" "jobspy" {
  api_id             = aws_apigatewayv2_api.api.id
  integration_type   = "AWS_PROXY"
  integration_uri    = aws_lambda_function.jobspy_fetcher.invoke_arn
  integration_method = "POST"
}

# User API routes
resource "aws_apigatewayv2_route" "register" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "POST /api/register"
  target    = "integrations/${aws_apigatewayv2_integration.api.id}"
}

resource "aws_apigatewayv2_route" "login" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "POST /api/login"
  target    = "integrations/${aws_apigatewayv2_integration.api.id}"
}

# Jobs API routes
resource "aws_apigatewayv2_route" "jobs" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "GET /api/jobs"
  target    = "integrations/${aws_apigatewayv2_integration.jobs_api.id}"
}

resource "aws_apigatewayv2_route" "notifications" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "GET /api/notifications"
  target    = "integrations/${aws_apigatewayv2_integration.jobs_api.id}"
}

# JobSpy route
resource "aws_apigatewayv2_route" "jobspy" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "POST /jobs/fetch"
  target    = "integrations/${aws_apigatewayv2_integration.jobspy.id}"
}

# Catch-all route for User API (for protected routes)
resource "aws_apigatewayv2_route" "api_default" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.api.id}"
}

# Deploy to stage
resource "aws_apigatewayv2_stage" "api" {
  api_id      = aws_apigatewayv2_api.api.id
  name        = "$default"
  auto_deploy = true
}

# Allow API Gateway to invoke Lambdas
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*/*"
}

resource "aws_lambda_permission" "jobs_api_gateway" {
  statement_id  = "AllowAPIGatewayInvokeJobs"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.jobs_api.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*/*"
}

resource "aws_lambda_permission" "jobspy_api_gateway" {
  statement_id  = "AllowAPIGatewayInvokeJobspy"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.jobspy_fetcher.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*/*"
}

output "api_url" {
  value = aws_apigatewayv2_api.api.api_endpoint
}
