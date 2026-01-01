# SQS & ECS Fargate Implementation Plan

## 📚 Part 1: Understanding SQS (Simple Queue Service)

### What is SQS?

**Amazon Simple Queue Service (SQS)** is like a **digital post office box** for messages between different parts of your application. Instead of one service directly talking to another, they communicate by putting messages into a queue and reading messages from that queue.

### Why Use SQS?

Think of your job processing system like a restaurant:

**Without SQS (Direct Communication):**
```
Customer → Waiter → Kitchen → Waiter → Customer
```
- If the kitchen is busy, the waiter has to wait
- If the kitchen crashes, everything stops
- You can't handle multiple orders efficiently

**With SQS (Queue-Based):**
```
Customer → Order Queue → [Multiple Kitchen Workers] → Completed Orders
```
- Customers can place orders even if kitchen is busy
- Kitchen workers process orders when ready
- If one worker crashes, others keep working
- Orders are never lost

### Real-World Analogy

Imagine you're processing job listings:

1. **Your Scanner** finds new jobs every 10 minutes
2. Instead of processing them immediately, it **puts them in a queue** (like a to-do list)
3. **Your Worker** reads from the queue and processes jobs when it's ready
4. If the worker is busy, jobs wait safely in the queue
5. If the worker crashes, jobs stay in the queue until it's back

### Key SQS Concepts

#### 1. **Queue**
A **queue** is a container that holds messages. Think of it like a mailbox:
- Messages are stored until someone reads them
- Messages are processed in order (or priority order)
- Messages can be read multiple times if needed

#### 2. **Message**
A **message** is a piece of data you want to send. In our case:
```json
{
  "job_id": "12345",
  "action": "process_job",
  "data": {
    "title": "Software Engineer",
    "company": "Tech Corp",
    "location": "San Francisco"
  }
}
```

#### 3. **Producer (Sender)**
The service that **puts messages** into the queue. In our system:
- **Scanner** = Producer (finds jobs and sends them to queue)

#### 4. **Consumer (Receiver)**
The service that **reads and processes** messages from the queue:
- **Worker/Matcher** = Consumer (reads jobs from queue and processes them)

#### 5. **Visibility Timeout**
When a consumer reads a message, it becomes "invisible" to others for a period of time:
- **Default: 30 seconds**
- **Purpose**: Gives the consumer time to process without other consumers grabbing the same message
- **If processing takes longer**: The message becomes visible again (so it can be retried)
- **After successful processing**: Consumer deletes the message

**Example:**
```
1. Worker reads message #1 → Message becomes invisible for 30 seconds
2. Worker processes message (takes 20 seconds) → Success!
3. Worker deletes message → Message is gone forever
```

**If processing fails:**
```
1. Worker reads message #1 → Message becomes invisible for 30 seconds
2. Worker crashes after 35 seconds → Message becomes visible again
3. Another worker reads message #1 → Retry!
```

#### 6. **Dead Letter Queue (DLQ)**
A **safety net** for messages that fail repeatedly:
- If a message fails processing **3 times** (configurable)
- It moves to the DLQ instead of being retried forever
- You can investigate failed messages later

**Why it matters:**
```
Bad message (invalid data) → Tries 3 times → Fails → Moves to DLQ
Good message (temporary error) → Tries 3 times → Eventually succeeds
```

### How SQS Works

#### Standard Queue (What We'll Use)
```
1. Producer sends message → Message arrives in queue
2. Consumer polls queue → Gets message
3. Message becomes invisible (visibility timeout)
4. Consumer processes message
5. Consumer deletes message → Done!
```

**Characteristics:**
- **At-least-once delivery**: Messages might be delivered more than once
- **Best-effort ordering**: Messages usually arrive in order, but not guaranteed
- **Unlimited throughput**: Can handle any volume
- **Nearly unlimited queue size**: Can store millions of messages

#### FIFO Queue (Alternative)
- **Exactly-once delivery**: Each message delivered exactly once
- **Strict ordering**: Messages always arrive in order
- **Limited throughput**: 3,000 messages/second with batching
- **Deduplication**: Prevents duplicate messages

**When to use FIFO?** When order matters critically (e.g., financial transactions)

**When to use Standard?** For most use cases, including ours (job processing)

### SQS Pricing

**Free Tier:**
- First 1 million requests per month: **FREE**

**After free tier:**
- $0.40 per million requests
- Messages stored: **FREE**
- Data transfer: First 1 GB/month **FREE**

**Example cost:**
- 1,000 jobs processed per day = ~30,000/month
- Cost: **$0.00** (within free tier!)

### SQS vs. Other Solutions

| Feature | SQS | RabbitMQ | Redis | Kafka |
|---------|-----|----------|-------|-------|
| Managed | ✅ Yes | ❌ No | ❌ No | ⚠️ Partial |
| Setup Complexity | ⭐ Easy | ⭐⭐⭐ Hard | ⭐⭐ Medium | ⭐⭐⭐ Very Hard |
| AWS Integration | ✅ Native | ❌ Manual | ❌ Manual | ⚠️ Partial |
| Cost | 💰 Very Low | 💰 Medium | 💰 Low | 💰 Medium |
| Best For | AWS workloads | Complex routing | Simple queues | High throughput |

---

## 📚 Part 2: Understanding ECS Fargate

### What is ECS Fargate?

**Amazon Elastic Container Service (ECS) Fargate** is AWS's way to run Docker containers **without managing servers**. It's like having a cloud computer that automatically handles:
- Server provisioning
- Server maintenance
- Scaling up/down
- Load balancing

### Why Use ECS Fargate?

Think of running containers like running a restaurant:

**Without Fargate (EC2/EKS):**
```
You = Restaurant Owner
- Rent the building (EC2 instances)
- Hire managers (install Kubernetes)
- Maintain equipment (server updates)
- Handle capacity (add/remove tables)
- Pay even when empty (24/7 server costs)
```

**With Fargate:**
```
AWS = Restaurant Owner
- Provides the building (infrastructure)
- Handles all management (auto-scaling, updates)
- You just provide the recipe (Docker container)
- Pay only when serving customers (per-task pricing)
```

### Real-World Analogy

**Your Worker Service:**
- Runs continuously, processing jobs from SQS
- Needs to handle varying workloads
- Should automatically scale up when busy
- Should be resilient (restarts if crashes)

**With Fargate:**
- You create a Docker container with your worker code
- Fargate runs it for you
- If it crashes, Fargate restarts it
- If you need more capacity, Fargate spins up more containers

### Key ECS Fargate Concepts

#### 1. **Task Definition**
A **blueprint** for your container:
```json
{
  "family": "job-worker",
  "containerDefinitions": [{
    "name": "worker",
    "image": "your-docker-image:latest",
    "memory": 512,
    "cpu": 256,
    "environment": [
      {"name": "SQS_QUEUE_URL", "value": "..."},
      {"name": "DATABASE_URL", "value": "..."}
    ]
  }]
}
```

**Think of it as:** A recipe card with all the ingredients (container image, memory, CPU, environment variables)

#### 2. **Task**
An **instance** of your task definition running:
- One task = One container running
- Each task processes messages independently
- You can run multiple tasks (multiple containers)

**Think of it as:** One actual dish cooked from the recipe

#### 3. **Service**
A **long-running group of tasks** that:
- Keeps a certain number of tasks running
- Restarts tasks if they crash
- Can scale up/down based on demand

**Think of it as:** The kitchen staff that keeps cooking (maintains running containers)

#### 4. **Cluster**
A **logical grouping** of your tasks and services:
- You can have multiple services in one cluster
- Cluster handles resource allocation
- You don't see or manage the underlying servers

**Think of it as:** The entire restaurant building (but you never see the building, just your kitchen)

#### 5. **Task Role (IAM)**
**Permissions** for your container to access AWS services:
- Read from SQS
- Write to database
- Send emails via SES
- Access secrets

**Think of it as:** An ID badge that lets your container access different parts of AWS

### How ECS Fargate Works

```
1. You create Task Definition → "Here's my container recipe"
2. You create Service → "Keep 2 containers running"
3. ECS Fargate:
   - Pulls your Docker image
   - Starts containers
   - Monitors health
   - Restarts if crashes
   - Scales if needed
4. Your containers run and process work
```

### Fargate vs. Lambda

| Feature | ECS Fargate | Lambda |
|---------|-------------|--------|
| **Best For** | Long-running processes | Short, event-driven tasks |
| **Runtime** | Up to hours/days | Max 15 minutes |
| **Cost Model** | Pay per task-second | Pay per invocation |
| **Cold Starts** | Minimal | Can be significant |
| **Control** | Full (choose OS, versions) | Limited (AWS runtime) |
| **SQS Integration** | Long polling | Event source mapping |

**For our use case (Worker processing SQS messages):**
- ✅ **ECS Fargate**: Better for continuous processing
- ❌ **Lambda**: Limited to 15 minutes, might timeout

### Fargate Pricing

**Charges:**
- **vCPU**: $0.04048 per vCPU-hour
- **Memory**: $0.004445 per GB-hour
- **Storage**: $0.20 per GB-month (for images)

**Example (1 task, 0.5 vCPU, 1GB RAM, running 24/7):**
- vCPU: 0.5 × $0.04048 × 730 hours = $14.78/month
- Memory: 1 × $0.004445 × 730 hours = $3.24/month
- **Total: ~$18/month** for one always-running worker

**Cost optimization:**
- Use minimum required resources
- Scale down when idle
- Use spot tasks for non-critical workloads

---

## 🏗️ Part 3: Architecture Overview

### Current Architecture (Before SQS/ECS)

```
┌─────────────┐
│   Scanner   │──┐
│  (Lambda)   │  │
└─────────────┘  │
                 │
┌─────────────┐  │
│   Worker    │  │  Direct Database Access
│  (Lambda)   │  │  (Tightly Coupled)
└─────────────┘  │
                 │
┌─────────────┐  │
│  PostgreSQL │◀─┘
└─────────────┘
```

**Problems:**
- Scanner and Worker tightly coupled
- If Worker is busy, Scanner blocks
- No retry mechanism
- Hard to scale independently

### New Architecture (With SQS/ECS)

```
┌─────────────┐
│   Scanner   │
│  (Lambda)   │
└──────┬──────┘
       │
       │ Sends messages
       ▼
┌─────────────┐
│ SQS Queue   │  ← Messages wait here safely
└──────┬──────┘
       │
       │ Long polling
       │ (multiple workers can read)
       ▼
┌─────────────────────────┐
│   ECS Fargate Service   │
│  ┌───────────────────┐  │
│  │  Worker Container │  │ ← Processes messages
│  └───────────────────┘  │
│  ┌───────────────────┐  │
│  │  Worker Container │  │ ← Auto-scales
│  └───────────────────┘  │
└─────────────┬───────────┘
              │
              │ Writes results
              ▼
       ┌─────────────┐
       │  PostgreSQL │
       └─────────────┘
```

**Benefits:**
- ✅ **Decoupled**: Scanner doesn't wait for Worker
- ✅ **Resilient**: Messages persist if Worker crashes
- ✅ **Scalable**: Can run multiple workers
- ✅ **Retry Logic**: Failed messages retry automatically
- ✅ **Monitoring**: Track queue depth, processing rate

### Message Flow

```
1. Scanner finds new jobs
   ↓
2. Scanner sends message to SQS:
   {
     "job_id": "123",
     "action": "process_job",
     "data": { ... }
   }
   ↓
3. Message sits in SQS queue
   ↓
4. Worker (ECS Fargate) polls queue
   ↓
5. Worker receives message (message becomes invisible)
   ↓
6. Worker processes:
   - Fetches job details
   - Analyzes with AI
   - Matches with user preferences
   - Saves to database
   ↓
7. Worker deletes message from queue
   ↓
8. If Worker crashes during step 6:
   - Message becomes visible again after timeout
   - Another worker picks it up
   - Retry!
```

---

## 📋 Part 4: Implementation Plan

### Phase 1: Infrastructure Setup (Terraform)

#### Step 1.1: Create SQS Queue

**File: `infra/terraform/sqs.tf`**

```hcl
# Main job processing queue
resource "aws_sqs_queue" "job_queue" {
  name                      = "jobping-job-queue"
  visibility_timeout_seconds = 300    # 5 minutes (match task timeout)
  message_retention_seconds  = 1209600 # 14 days (how long messages stay)
  receive_wait_time_seconds  = 20      # Long polling (wait for messages)
  
  # Enable server-side encryption
  kms_master_key_id = "alias/aws/sqs"
  kms_data_key_reuse_period_seconds = 300

  tags = {
    Name        = "jobping-job-queue"
    Environment = var.environment
  }
}

# Dead Letter Queue for failed messages
resource "aws_sqs_queue" "job_queue_dlq" {
  name                      = "jobping-job-queue-dlq"
  message_retention_seconds = 1209600 # 14 days

  tags = {
    Name        = "jobping-job-queue-dlq"
    Environment = var.environment
  }
}

# Connect DLQ to main queue
resource "aws_sqs_queue_redrive_policy" "job_queue" {
  queue_url = aws_sqs_queue.job_queue.id

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.job_queue_dlq.arn
    maxReceiveCount     = 3  # After 3 failed attempts, move to DLQ
  })
}

# CloudWatch alarms for monitoring
resource "aws_cloudwatch_metric_alarm" "queue_depth" {
  alarm_name          = "jobping-queue-depth"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 300  # 5 minutes
  statistic           = "Average"
  threshold           = 100  # Alert if more than 100 messages waiting
  alarm_description   = "Alert when queue has too many messages"
  
  dimensions = {
    QueueName = aws_sqs_queue.job_queue.name
  }

  alarm_actions = [] # Add SNS topic for notifications if needed
}

# Outputs
output "sqs_queue_url" {
  value       = aws_sqs_queue.job_queue.url
  description = "URL of the SQS queue"
}

output "sqs_queue_arn" {
  value       = aws_sqs_queue.job_queue.arn
  description = "ARN of the SQS queue"
}

output "sqs_dlq_url" {
  value       = aws_sqs_queue.job_queue_dlq.url
  description = "URL of the Dead Letter Queue"
}
```

#### Step 1.2: Create ECS Cluster

**File: `infra/terraform/ecs.tf`**

```hcl
# ECS Cluster (logical grouping, no servers to manage!)
resource "aws_ecs_cluster" "main" {
  name = "jobping-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"  # Better monitoring
  }

  tags = {
    Name        = "jobping-cluster"
    Environment = var.environment
  }
}

# CloudWatch Log Group for worker logs
resource "aws_cloudwatch_log_group" "worker" {
  name              = "/ecs/jobping-worker"
  retention_in_days = 14

  tags = {
    Name        = "jobping-worker-logs"
    Environment = var.environment
  }
}

# IAM Role for ECS Tasks (what permissions containers have)
resource "aws_iam_role" "ecs_task_role" {
  name = "jobping-ecs-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })

  tags = {
    Name = "jobping-ecs-task-role"
  }
}

# IAM Role for ECS Task Execution (pulling images, writing logs)
resource "aws_iam_role" "ecs_execution_role" {
  name = "jobping-ecs-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })

  tags = {
    Name = "jobping-ecs-execution-role"
  }
}

# Attach policies to execution role
resource "aws_iam_role_policy_attachment" "ecs_execution" {
  role       = aws_iam_role.ecs_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Attach policies to task role (what the container can do)
resource "aws_iam_role_policy" "task_sqs" {
  name = "jobping-task-sqs-policy"
  role = aws_iam_role.ecs_task_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:GetQueueUrl"
        ]
        Resource = [
          aws_sqs_queue.job_queue.arn,
          aws_sqs_queue.job_queue_dlq.arn
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy" "task_database" {
  name = "jobping-task-database-policy"
  role = aws_iam_role.ecs_task_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "rds-db:connect"
        ]
        Resource = "arn:aws:rds-db:${var.aws_region}:*:dbuser:${aws_db_instance.postgres.id}/*"
      }
    ]
  })
}

# Task Definition (blueprint for containers)
resource "aws_ecs_task_definition" "worker" {
  family                   = "jobping-worker"
  requires_compatibilities = ["FARGATE"]
  network_mode            = "awsvpc"
  cpu                     = 512   # 0.5 vCPU
  memory                  = 1024  # 1 GB RAM
  execution_role_arn      = aws_iam_role.ecs_execution_role.arn
  task_role_arn           = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode([{
    name  = "worker"
    image = "${var.aws_account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/jobping-worker:latest"
    
    essential = true

    environment = [
      {
        name  = "ENVIRONMENT"
        value = var.environment
      },
      {
        name  = "SQS_QUEUE_URL"
        value = aws_sqs_queue.job_queue.url
      },
      {
        name  = "DATABASE_URL"
        value = "postgres://jobscanner:${var.db_password}@${aws_db_instance.postgres.endpoint}/jobscanner?sslmode=require"
      }
    ]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.worker.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "worker"
      }
    }

    # Health check
    healthCheck = {
      command     = ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 60
    }
  }])

  tags = {
    Name        = "jobping-worker-task"
    Environment = var.environment
  }
}

# ECS Service (keeps tasks running)
resource "aws_ecs_service" "worker" {
  name            = "jobping-worker-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.worker.arn
  desired_count   = 2  # Run 2 containers
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = data.aws_subnets.default.ids
    assign_public_ip = true  # For pulling images and accessing internet
    security_groups  = [aws_security_group.ecs_tasks.id]
  }

  # Auto-scaling (optional)
  # Will cover this in Phase 3

  tags = {
    Name        = "jobping-worker-service"
    Environment = var.environment
  }
}

# Security group for ECS tasks
resource "aws_security_group" "ecs_tasks" {
  name        = "jobping-ecs-tasks"
  description = "Security group for ECS Fargate tasks"
  vpc_id      = data.aws_vpc.default.id

  # Allow outbound to RDS
  egress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    security_groups = [aws_security_group.rds.id]
    description = "Allow connection to RDS"
  }

  # Allow all outbound (for SQS, CloudWatch, etc.)
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }

  tags = {
    Name = "jobping-ecs-tasks-sg"
  }
}

# Update RDS security group to allow ECS tasks
resource "aws_security_group_rule" "rds_allow_ecs" {
  type                     = "ingress"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.ecs_tasks.id
  security_group_id        = aws_security_group.rds.id
  description              = "Allow ECS tasks to connect to RDS"
}
```

#### Step 1.3: Create ECR Repository (for Docker images)

**File: `infra/terraform/ecr.tf`**

```hcl
# ECR Repository for worker Docker image
resource "aws_ecr_repository" "worker" {
  name                 = "jobping-worker"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true  # Scan for vulnerabilities
  }

  tags = {
    Name        = "jobping-worker-ecr"
    Environment = var.environment
  }
}

# Output ECR repository URL
output "ecr_repository_url" {
  value       = aws_ecr_repository.worker.repository_url
  description = "URL of the ECR repository"
}
```

#### Step 1.4: Add Variable for AWS Account ID

**Update: `infra/terraform/main.tf`**

```hcl
variable "aws_account_id" {
  description = "AWS Account ID"
  type        = string
}
```

### Phase 2: Backend Code Implementation

#### Step 2.1: Add AWS SDK Dependencies

**Update: `backend/go.mod`**

```bash
cd backend
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/sqs
go get github.com/aws/aws-sdk-go-v2/credentials
```

#### Step 2.2: Create SQS Client Infrastructure

**File: `backend/internal/infrastructure/sqs/client.go`**

```go
package sqs

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Client interface for SQS operations
type Client interface {
	SendMessage(ctx context.Context, body interface{}) error
	ReceiveMessages(ctx context.Context, maxMessages int32) ([]Message, error)
	DeleteMessage(ctx context.Context, receiptHandle string) error
}

// Message represents an SQS message
type Message struct {
	Body          string
	ReceiptHandle string
	MessageID     string
}

type sqsClient struct {
	client   *sqs.Client
	queueURL string
}

// NewClient creates a new SQS client
func NewClient(queueURL string) (Client, error) {
	// Check if local (LocalStack)
	if os.Getenv("ENVIRONMENT") == "local" {
		return NewLocalClient(queueURL)
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	return &sqsClient{
		client:   sqs.NewFromConfig(cfg),
		queueURL: queueURL,
	}, nil
}

// SendMessage sends a message to the SQS queue
func (c *sqsClient) SendMessage(ctx context.Context, body interface{}) error {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return err
	}

	_, err = c.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(string(bodyJSON)),
	})

	return err
}

// ReceiveMessages receives messages from the SQS queue (long polling)
func (c *sqsClient) ReceiveMessages(ctx context.Context, maxMessages int32) ([]Message, error) {
	result, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.queueURL),
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     20, // Long polling (wait up to 20 seconds)
	})

	if err != nil {
		return nil, err
	}

	messages := make([]Message, len(result.Messages))
	for i, msg := range result.Messages {
		messages[i] = Message{
			Body:          aws.ToString(msg.Body),
			ReceiptHandle: aws.ToString(msg.ReceiptHandle),
			MessageID:     aws.ToString(msg.MessageId),
		}
	}

	return messages, nil
}

// DeleteMessage deletes a message from the queue after successful processing
func (c *sqsClient) DeleteMessage(ctx context.Context, receiptHandle string) error {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})

	return err
}
```

**File: `backend/internal/infrastructure/sqs/local_client.go`**

```go
package sqs

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// NewLocalClient creates an SQS client for LocalStack (local development)
func NewLocalClient(queueURL string) (Client, error) {
	endpoint := "http://localhost:4566"
	if ep := os.Getenv("LOCALSTACK_ENDPOINT"); ep != "" {
		endpoint = ep
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           endpoint,
					SigningRegion: region,
				}, nil
			})),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		return nil, err
	}

	return &sqsClient{
		client:   sqs.NewFromConfig(cfg),
		queueURL: queueURL,
	}, nil
}
```

#### Step 2.3: Update Config

**File: `backend/internal/config/config.go`**

```go
type Config struct {
	Environment string
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTExpiry   int
	SQSQueueURL string  // Add this
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Environment: getEnv("ENVIRONMENT", "local"),
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTExpiry:   getEnvInt("JWT_EXPIRY_HOURS", 24),
		SQSQueueURL: getEnv("SQS_QUEUE_URL", ""),  // Add this
	}
}
```

#### Step 2.4: Create Job Queue Service

**File: `backend/internal/features/job_queue/service/job_queue_service.go`**

```go
package service

import (
	"context"
	"encoding/json"

	"github.com/jobping/backend/internal/infrastructure/sqs"
)

// JobMessage represents a message in the job queue
type JobMessage struct {
	JobID  string          `json:"job_id"`
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

// JobQueueService handles job queue operations
type JobQueueService struct {
	sqsClient sqs.Client
}

// NewJobQueueService creates a new job queue service
func NewJobQueueService(sqsClient sqs.Client) *JobQueueService {
	return &JobQueueService{
		sqsClient: sqsClient,
	}
}

// EnqueueJob adds a job to the queue
func (s *JobQueueService) EnqueueJob(ctx context.Context, jobID string, action string, data interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	msg := JobMessage{
		JobID:  jobID,
		Action: action,
		Data:   dataJSON,
	}

	return s.sqsClient.SendMessage(ctx, msg)
}

// ProcessJobs processes messages from the queue
// handler function processes each message and returns error if processing failed
func (s *JobQueueService) ProcessJobs(ctx context.Context, handler func(context.Context, JobMessage) error) error {
	messages, err := s.sqsClient.ReceiveMessages(ctx, 10) // Get up to 10 messages
	if err != nil {
		return err
	}

	for _, msg := range messages {
		var jobMsg JobMessage
		if err := json.Unmarshal([]byte(msg.Body), &jobMsg); err != nil {
			// Invalid message format - delete it
			_ = s.sqsClient.DeleteMessage(ctx, msg.ReceiptHandle)
			continue
		}

		// Process the message
		if err := handler(ctx, jobMsg); err != nil {
			// Processing failed - don't delete, let it retry
			// Message will become visible again after visibility timeout
			continue
		}

		// Success - delete message from queue
		if err := s.sqsClient.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
			return err
		}
	}

	return nil
}
```

#### Step 2.5: Update Scanner to Send to SQS

**File: `backend/cmd/scanner/main.go` (example structure)**

```go
package main

import (
	"context"
	"log"

	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/infrastructure/sqs"
	"github.com/jobping/backend/internal/features/job_queue/service"
)

func main() {
	cfg := config.Load()

	// Initialize SQS client
	sqsClient, err := sqs.NewClient(cfg.SQSQueueURL)
	if err != nil {
		log.Fatalf("Failed to create SQS client: %v", err)
	}

	queueService := service.NewJobQueueService(sqsClient)

	// Find new jobs (your existing scanner logic)
	newJobs := findNewJobs() // Your existing function

	// Send each job to SQS queue
	ctx := context.Background()
	for _, job := range newJobs {
		err := queueService.EnqueueJob(ctx, job.ID, "process_job", map[string]interface{}{
			"job_id":   job.ID,
			"title":    job.Title,
			"company":  job.Company,
			"location": job.Location,
		})
		if err != nil {
			log.Printf("Failed to enqueue job %s: %v", job.ID, err)
			continue
		}
		log.Printf("Enqueued job %s", job.ID)
	}
}
```

#### Step 2.6: Create Worker Command

**File: `backend/cmd/worker/main.go`**

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
	"github.com/jobping/backend/internal/infrastructure/sqs"
	"github.com/jobping/backend/internal/features/job_queue/service"
	// Import your job processing services here
)

func main() {
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize SQS client
	sqsClient, err := sqs.NewClient(cfg.SQSQueueURL)
	if err != nil {
		log.Fatalf("Failed to create SQS client: %v", err)
	}

	queueService := service.NewJobQueueService(sqsClient)

	// Initialize your job processing services
	// jobService := service.NewJobService(...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down worker...")
		cancel()
	}()

	log.Println("Worker started, polling for messages...")

	// Main processing loop
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped")
			return
		default:
			// Process messages from queue
			if err := queueService.ProcessJobs(ctx, processJob); err != nil {
				log.Printf("Error processing jobs: %v", err)
				// Wait before retrying
				time.Sleep(5 * time.Second)
			}
			// Small delay between polling cycles
			time.Sleep(1 * time.Second)
		}
	}
}

// processJob processes a single job message
func processJob(ctx context.Context, msg service.JobMessage) error {
	log.Printf("Processing job: %s, action: %s", msg.JobID, msg.Action)

	// Parse message data
	var jobData map[string]interface{}
	if err := json.Unmarshal(msg.Data, &jobData); err != nil {
		return fmt.Errorf("invalid job data: %w", err)
	}

	// Process based on action
	switch msg.Action {
	case "process_job":
		// Your job processing logic here:
		// 1. Fetch job details
		// 2. Analyze with AI
		// 3. Match with user preferences
		// 4. Save to database
		// 5. Send notifications if matches found
		
		log.Printf("Processed job %s successfully", msg.JobID)
		return nil

	default:
		log.Printf("Unknown action: %s", msg.Action)
		return fmt.Errorf("unknown action: %s", msg.Action)
	}
}
```

#### Step 2.7: Update App Build to Include SQS

**File: `backend/internal/app/app.go`**

```go
// Add SQS client initialization if needed for API endpoints
// For now, scanner and worker use it directly
```

### Phase 3: Docker Setup

#### Step 3.1: Create Worker Dockerfile

**File: `backend/Dockerfile.worker`**

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build worker binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o worker \
    ./cmd/worker

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/worker .

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Run worker
CMD ["./worker"]
```

#### Step 3.2: Build and Push Docker Image

**File: `scripts/build-worker-docker.sh`**

```bash
#!/bin/bash
set -e

echo "Building worker Docker image..."

# Get AWS account ID and region from Terraform
cd infra/terraform
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
AWS_REGION=$(terraform output -raw aws_region 2>/dev/null || echo "us-east-1")
ECR_REPO=$(terraform output -raw ecr_repository_url 2>/dev/null)

if [ -z "$ECR_REPO" ]; then
  ECR_REPO="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/jobping-worker"
fi

cd ../..

# Build Docker image
docker build -f backend/Dockerfile.worker -t jobping-worker:latest backend/

# Tag for ECR
docker tag jobping-worker:latest ${ECR_REPO}:latest

# Login to ECR
aws ecr get-login-password --region ${AWS_REGION} | \
  docker login --username AWS --password-stdin ${ECR_REPO}

# Push to ECR
docker push ${ECR_REPO}:latest

echo "✓ Worker image pushed to ${ECR_REPO}:latest"
```

### Phase 4: Local Testing

#### Step 4.1: Add LocalStack to docker-compose

**Update: `docker-compose.yml`**

```yaml
services:
  # ... existing services ...

  localstack:
    image: localstack/localstack:latest
    container_name: jobping-localstack
    ports:
      - "4566:4566"
    environment:
      - SERVICES=sqs
      - DEBUG=1
      - DATA_DIR=/tmp/localstack/data
    volumes:
      - localstack_data:/tmp/localstack
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:4566/_localstack/health"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  localstack_data:
```

#### Step 4.2: Setup Script for LocalStack

**File: `scripts/setup-localstack.sh`**

```bash
#!/bin/bash

echo "Setting up LocalStack SQS queue..."

# Create queue
aws --endpoint-url=http://localhost:4566 sqs create-queue \
  --queue-name jobping-job-queue \
  --region us-east-1

# Get queue URL
QUEUE_URL=$(aws --endpoint-url=http://localhost:4566 sqs get-queue-url \
  --queue-name jobping-job-queue \
  --region us-east-1 \
  --query 'QueueUrl' \
  --output text)

echo "Queue created!"
echo "Queue URL: $QUEUE_URL"
echo ""
echo "Add this to your .env file:"
echo "SQS_QUEUE_URL=$QUEUE_URL"
```

#### Step 4.3: Test Locally

```bash
# Terminal 1: Start services
docker-compose up -d
bash scripts/setup-localstack.sh

# Terminal 2: Run scanner (sends messages to queue)
cd backend
export SQS_QUEUE_URL=http://localhost:4566/000000000000/jobping-job-queue
go run cmd/scanner/main.go

# Terminal 3: Run worker (processes messages from queue)
cd backend
export SQS_QUEUE_URL=http://localhost:4566/000000000000/jobping-job-queue
export DATABASE_URL=postgres://jobscanner:password@localhost:5433/jobscanner?sslmode=disable
go run cmd/worker/main.go
```

### Phase 5: Deployment Steps

#### Step 5.1: Deploy Infrastructure

```bash
cd infra/terraform

# Initialize (if first time)
terraform init

# Review changes
terraform plan

# Apply infrastructure
terraform apply

# Save outputs
terraform output -json > outputs.json
```

#### Step 5.2: Build and Push Docker Image

```bash
# Build and push worker image
bash scripts/build-worker-docker.sh
```

#### Step 5.3: Update ECS Service (if task definition changed)

```bash
# Force new deployment (picks up new image)
aws ecs update-service \
  --cluster jobping-cluster \
  --service jobping-worker-service \
  --force-new-deployment
```

#### Step 5.4: Monitor Deployment

```bash
# Watch ECS service
aws ecs describe-services \
  --cluster jobping-cluster \
  --services jobping-worker-service

# Watch logs
aws logs tail /ecs/jobping-worker --follow

# Watch queue
aws sqs get-queue-attributes \
  --queue-url <QUEUE_URL> \
  --attribute-names All
```

### Phase 6: Monitoring & Observability

#### CloudWatch Metrics to Monitor

1. **SQS Queue Depth**: `ApproximateNumberOfMessagesVisible`
   - Alert if > 100 messages waiting

2. **SQS Message Age**: `ApproximateAgeOfOldestMessage`
   - Alert if messages older than 5 minutes

3. **ECS Task Count**: `RunningTaskCount`
   - Should match desired count

4. **ECS CPU Utilization**: `CPUUtilization`
   - Alert if consistently > 80%

5. **ECS Memory Utilization**: `MemoryUtilization`
   - Alert if consistently > 80%

#### Create CloudWatch Dashboard

```bash
# Use AWS Console or Terraform to create dashboard
# Monitor:
# - Queue depth over time
# - Messages processed per minute
# - Task count
# - Error rate
```

---

## 🎯 Summary: What We're Building

### Before (Current)
- Scanner and Worker tightly coupled
- No retry mechanism
- Hard to scale

### After (With SQS + ECS Fargate)
- ✅ **Decoupled**: Scanner sends to queue, doesn't wait
- ✅ **Resilient**: Messages persist, auto-retry on failure
- ✅ **Scalable**: Run multiple workers, auto-scale based on queue depth
- ✅ **Observable**: Monitor queue depth, processing rate, errors
- ✅ **Cost-effective**: Pay only for what you use

### Cost Estimate

**SQS:**
- First 1M requests/month: FREE
- Your usage: ~100K/month → **$0.00**

**ECS Fargate:**
- 2 tasks × 0.5 vCPU × 1GB RAM
- Running 24/7: ~$36/month
- With auto-scaling (0-4 tasks): ~$0-72/month depending on load

**Total: ~$36/month** for a robust, scalable job processing system!

---

## 🚀 Next Steps

1. **Review this plan** and understand the concepts
2. **Implement Phase 1** (Terraform infrastructure)
3. **Test locally** with LocalStack
4. **Implement Phase 2** (Backend code)
5. **Test end-to-end** locally
6. **Deploy to AWS** (Phase 5)
7. **Monitor and optimize** (Phase 6)

---

## 📖 Additional Resources

- [AWS SQS Documentation](https://docs.aws.amazon.com/sqs/)
- [ECS Fargate Documentation](https://docs.aws.amazon.com/ecs/latest/developerguide/AWS_Fargate.html)
- [LocalStack Documentation](https://docs.localstack.cloud/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)

---

**Questions?** Review the concepts sections again, or check AWS documentation for detailed examples.

