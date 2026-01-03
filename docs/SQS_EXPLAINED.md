# SQS Explained for Beginners

## What is SQS? (Simple Analogy)

Think of SQS (Simple Queue Service) like a post office mailbox:

- **Sender** (Python Lambda) puts letters (messages) into the mailbox (queue)
- Messages wait in the mailbox until someone picks them up
- **Receiver** (Go Lambda) comes by, picks up letters, and processes them
- If something goes wrong, the letter goes to a "dead letter" mailbox (DLQ)

**Key Benefit**: The sender and receiver don't need to talk directly - they communicate through the mailbox!

---

## Why Use SQS?

### Problem Without SQS:
```
Python Lambda → Go Lambda (direct call)
     ↓
❌ If Go Lambda is busy/crashes → Python Lambda waits/hangs
❌ If Python Lambda sends 100 jobs → Go Lambda gets overwhelmed
❌ Hard to scale independently
```

### Solution With SQS:
```
Python Lambda → SQS Queue → Go Lambda
     ↓              ↓            ↓
✅ Sends message  ✅ Messages  ✅ Processes
   and returns      wait here    when ready
   immediately      safely
```

**Benefits:**
- **Decoupling**: Services don't need to know about each other
- **Reliability**: Messages are stored safely until processed
- **Scalability**: Can process messages at your own pace
- **Retry logic**: Failed messages can be retried automatically

---

## Key SQS Concepts

### 1. **Queue**
A container that holds messages. Like a mailbox with a specific address.

### 2. **Message**
Data (as JSON text) that you want to pass between services.

### 3. **Producer (Sender)**
Service that puts messages into the queue.

### 4. **Consumer (Receiver)**
Service that reads messages from the queue.

### 5. **Dead Letter Queue (DLQ)**
Where messages go if they fail too many times (max 3 tries in our case).

### 6. **Visibility Timeout**
When a message is read, it becomes "invisible" for 60 seconds. If processing succeeds, you delete it. If it fails, it becomes visible again for retry.

---

## How SQS Works in Your Codebase

### Architecture Flow

```
┌─────────────────┐         ┌──────────────┐         ┌──────────────┐
│ Python Lambda   │         │  SQS Queue   │         │  Go Lambda   │
│ (JobSpy)        │         │              │         │ (AI Filter)  │
└─────────────────┘         └──────────────┘         └──────────────┘
       │                           │                          │
       │ 1. Scrape 5 jobs          │                          │
       │                           │                          │
       │ 2. Send 5 messages ──────▶│                          │
       │    (each job = 1 msg)     │                          │
       │                           │                          │
       │ 3. Return immediately     │                          │
       │    ✅ Done!               │                          │
       │                           │                          │
       │                           │ 4. AWS automatically      │
       │                           │    triggers Go Lambda    │
       │                           │    when messages arrive  │
       │                           ├─────────────────────────▶│
       │                           │                          │
       │                           │ 5. Go Lambda receives    │
       │                           │    messages (batch)      │
       │                           │                          │
       │                           │                          │ 6. Process each job:
       │                           │                          │    - Run AI analysis
       │                           │                          │    - Save to database
       │                           │                          │    - Delete message ✅
       │                           │                          │
```

---

## Code Locations

### 1. **SQS Queue Definition** (Infrastructure)

**File**: `infra/terraform/sqs.tf`

```hcl
# This creates the actual SQS queue in AWS
resource "aws_sqs_queue" "jobs_to_filter" {
  name = "jobping-jobs-to-filter"
  visibility_timeout_seconds = 60  # Message invisible for 60s after reading
  message_retention_seconds  = 86400  # Keep messages for 1 day
}
```

**What it does**: Creates the queue when you run `terraform apply`.

---

### 2. **Sending Messages** (Python Lambda - Producer)

**File**: `python_workers/jobspy_fetcher/handler.py`

```python
# Initialize SQS client
sqs = boto3.client("sqs", region_name="us-east-1")
SQS_QUEUE_URL = os.environ.get("SQS_QUEUE_URL")

# Inside the handler function:
for _, job in jobs_df.iterrows():
    job_message = {
        "title": job.get("title"),
        "company": job.get("company"),
        # ... other job fields
    }
    
    # Send message to SQS queue
    sqs.send_message(
        QueueUrl=SQS_QUEUE_URL,
        MessageBody=json.dumps(job_message)  # Convert to JSON string
    )
```

**What it does**: 
- Scrapes jobs from job boards
- Converts each job to JSON
- Sends each job as a separate message to the queue
- Returns immediately (doesn't wait for processing)

---

### 3. **Receiving Messages** (Go Lambda - Consumer)

**File**: `backend/internal/features/job/handler/sqs.go`

```go
func (h *SQSHandler) HandleSQSEvent(ctx context.Context, sqsEvent events.SQSEvent) error {
    // AWS automatically batches messages
    for _, record := range sqsEvent.Records {
        // Parse the message (JSON string → Go struct)
        var jobInput service.JobInput
        json.Unmarshal([]byte(record.Body), &jobInput)
        
        // Process the job (AI analysis + save to DB)
        job, err := h.service.ProcessJob(ctx, &jobInput)
        if err != nil {
            // If it fails, message will become visible again for retry
            continue
        }
        
        // If successful, AWS automatically deletes the message
    }
    return nil
}
```

**What it does**:
- AWS Lambda automatically triggers this when messages arrive
- Receives messages in batches (1 at a time in our config)
- Processes each job (runs AI analysis, saves to database)
- If successful, AWS deletes the message
- If it fails 3 times, message goes to Dead Letter Queue

---

### 4. **Lambda Trigger Configuration**

**File**: `infra/terraform/sqs.tf` (line 68-74)

```hcl
# This connects SQS to Go Lambda
resource "aws_lambda_event_source_mapping" "sqs_to_go_lambda" {
  event_source_arn = aws_sqs_queue.jobs_to_filter.arn  # Which queue
  function_name    = aws_lambda_function.api.arn        # Which Lambda
  batch_size       = 1                                  # Process 1 at a time
  enabled          = true
}
```

**What it does**: 
- Tells AWS: "When messages arrive in this queue, automatically trigger this Lambda function"
- This is automatic - you don't need to write code to poll the queue!

---

## Real Example: What Happens When User Clicks "Fetch Jobs"

### Step-by-Step:

1. **User clicks button** → Frontend calls `POST /jobs/fetch`

2. **API Gateway** → Routes to Python Lambda

3. **Python Lambda** (`handler.py`):
   ```python
   # Scrapes 5 jobs from Indeed
   jobs = scrape_jobs(...)  # Returns 5 jobs
   
   # For each job, send to SQS:
   for job in jobs:
       sqs.send_message(
           QueueUrl="https://sqs.us-east-1.amazonaws.com/123/jobping-jobs-to-filter",
           MessageBody='{"title": "Software Engineer", "company": "Google", ...}'
       )
   
   # Returns immediately to frontend:
   return {"jobs_queued": 5}
   ```

4. **SQS Queue**:
   - Now contains 5 messages
   - Each message is a JSON string with one job

5. **AWS Automatically**:
   - Sees 5 messages in queue
   - Triggers Go Lambda 5 times (batch_size=1)
   - Each trigger has 1 message

6. **Go Lambda** (`sqs.go`):
   ```go
   // For each message:
   // 1. Parse JSON
   job := parseMessage(record.Body)
   
   // 2. Run AI analysis
   aiResult := aiClient.AnalyzeJob(job.Title, job.Company, job.Description)
   
   // 3. Save to database
   db.Save(job, aiResult)
   
   // 4. Message automatically deleted by AWS ✅
   ```

7. **Result**: All 5 jobs are now in database with AI analysis!

---

## Important SQS Features Used

### 1. **Visibility Timeout (60 seconds)**
- When Go Lambda reads a message, it becomes "invisible"
- If processing takes < 60s and succeeds → message deleted
- If processing fails/crashes → message becomes visible again → retry

### 2. **Dead Letter Queue (DLQ)**
- If a message fails 3 times, it goes to DLQ
- You can inspect failed messages later
- Prevents infinite retry loops

### 3. **Long Polling (20 seconds)**
- Lambda waits up to 20 seconds for messages
- More efficient than constantly checking
- Reduces AWS API calls

---

## Local Development (LocalStack)

For local testing, we use **LocalStack** - a tool that emulates AWS services:

**File**: `docker-compose.yml`

```yaml
localstack:
  image: localstack/localstack:latest
  ports:
    - "4566:4566"  # LocalStack SQS endpoint
```

**Usage**:
```python
# Instead of real AWS SQS:
SQS_QUEUE_URL = "https://sqs.us-east-1.amazonaws.com/123/queue"

# Use LocalStack:
SQS_QUEUE_URL = "http://localhost:4566/000000000000/jobping-jobs-to-filter"
```

The code works the same, but uses LocalStack instead of real AWS!

---

## Summary

**SQS = Message Queue Service**

- **Producer** (Python) sends messages → Queue
- Messages wait safely in queue
- **Consumer** (Go) processes messages when ready
- Automatic retry on failure
- Dead letter queue for failed messages
- No direct connection needed between services

**Benefits in Your App:**
- Python Lambda can send 100 jobs instantly (returns fast)
- Go Lambda processes them at its own pace
- If Go Lambda crashes, messages wait safely
- Can scale each service independently
- Reliable message delivery

---

## Common Questions

**Q: Why not call Go Lambda directly from Python?**  
A: If Go Lambda is slow/busy, Python would wait. With SQS, Python sends and returns immediately.

**Q: What if Go Lambda crashes while processing?**  
A: Message becomes visible again after 60 seconds → automatic retry (up to 3 times).

**Q: Can multiple Go Lambdas process messages?**  
A: Yes! AWS can automatically scale Lambda based on queue depth. Multiple instances can process messages in parallel.

**Q: How fast is SQS?**  
A: Very fast - messages are delivered within seconds, often instantly.

---

## Further Reading

- [AWS SQS Documentation](https://docs.aws.amazon.com/sqs/)
- [SQS Best Practices](https://docs.aws.amazon.com/sqs/latest/dg/sqs-best-practices.html)


