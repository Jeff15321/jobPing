# SQS Queue for job processing
resource "aws_sqs_queue" "job_queue" {
  name                      = "jobscanner-job-queue"
  delay_seconds             = 0
  max_message_size          = 262144
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
  
  tags = {
    Name        = "jobscanner-job-queue"
    Environment = var.environment
  }
}

# Dead Letter Queue
resource "aws_sqs_queue" "job_queue_dlq" {
  name = "jobscanner-job-queue-dlq"
  
  tags = {
    Name        = "jobscanner-job-queue-dlq"
    Environment = var.environment
  }
}

output "sqs_queue_url" {
  value = aws_sqs_queue.job_queue.url
}
