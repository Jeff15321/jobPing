# Job Feature Documentation

## Overview

The `job` feature provides **CRUD operations** for jobs stored in the database. This is the core data model for job postings. It also includes a **legacy service** (`JobService`) that performs end-to-end job processing for **local development/testing only**.

**⚠️ Important**: In production, job processing is handled by the 4-stage SQS pipeline (`job_analysis`, `user_fanout`, `user_analysis`, `notification` features). The `JobService.ProcessJob` method is only used for local testing.

## File Structure

```
backend/internal/features/job/
├── handler/
│   ├── dto.go              # Data Transfer Objects (response models)
│   ├── http.go             # HTTP handlers for job endpoints
│   ├── job_analysis.go     # ⚠️ OLD - Not used, should be removed
│   ├── mock.go             # Mock job fetching for local dev
│   └── sqs.go              # ⚠️ OLD - Not used, should be removed
├── joberr/
│   └── errors.go           # Job-specific error definitions
├── model/
│   └── job.go              # Job domain model
├── module.go               # Route registration
├── repository/
│   └── job_repository.go   # Database operations for jobs
└── service/
    ├── ai_client.go        # AI client interface (legacy - used by JobService)
    ├── job_service.go      # ⚠️ Legacy service for local dev only
    └── sqs_client.go       # ⚠️ Legacy SQS client (uses old NOTIFY_SQS_QUEUE_URL)
```

## Components

### Model (`model/job.go`)

**Purpose**: Defines the `Job` domain model.

**Key Fields**:
- `ID`, `Title`, `Company`, `Location`, `JobURL`, `Description`
- `JobType`, `IsRemote`, `MinSalary`, `MaxSalary`, `DatePosted`
- `AIScore`, `AIAnalysis` - Legacy AI analysis fields (not used in new pipeline)
- `CompanyInfo` - Company research data (JSONB)
- `CompanyInfoUpdatedAt` - Timestamp for company info freshness check
- `Status` - `pending`, `processed`, or `failed`
- `CreatedAt`, `UpdatedAt` - Timestamps

**Usage**: Used by repository and service layers to represent job data.

---

### Repository (`repository/job_repository.go`)

**Purpose**: Database operations for jobs (CRUD).

**Interface**:
```go
type JobRepository interface {
    Create(ctx, job) error
    GetByID(ctx, id) (*Job, error)
    GetByURL(ctx, url) (*Job, error)
    GetAll(ctx, limit) ([]Job, error)
    GetProcessed(ctx, limit) ([]Job, error)
    Update(ctx, job) error
    UpdateCompanyInfo(ctx, id, companyInfo) error
    ExistsByURL(ctx, url) (bool, error)
    IsCompanyInfoFresh(ctx, id) (bool, error)  // Checks if < 6 months old
}
```

**Key Methods**:
- `Create` - Insert new job
- `GetByID` - Fetch job by UUID
- `GetByURL` - Fetch job by URL (for duplicate detection)
- `GetProcessed` - Fetch processed jobs for display
- `UpdateCompanyInfo` - Update company research data and timestamp
- `IsCompanyInfoFresh` - Check if company info is less than 6 months old

**Database Table**: `jobs`

**Usage**: Used by all features that need to read/write job data.

---

### Service (`service/job_service.go`)

**⚠️ LEGACY - Local Development Only**

**Purpose**: End-to-end job processing for local testing. In production, this logic is split across the 4-stage pipeline.

**Key Methods**:
- `ProcessJob(ctx, input)` - Processes a job:
  1. Checks if job already exists
  2. Researches company (if needed)
  3. Runs AI analysis
  4. Saves to database
  5. Matches to all users
  6. Queues notifications (if threshold met)

- `GetJobs(ctx, limit)` - Returns processed jobs for display

**Dependencies**:
- `JobRepository` - Database operations
- `AIClient` - AI analysis (legacy interface)
- `UserRepository` - User matching
- `UserJobMatchRepository` - Match storage
- `SQSClient` - Notification queuing (legacy)

**⚠️ Issues**:
- Contains logic that should be in separate features
- Uses old `NOTIFY_SQS_QUEUE_URL` env var
- Still used by `mock.go` and `http.go` for local testing

**Recommendation**: Keep for local dev, but mark as deprecated. Consider simplifying to just CRUD operations.

---

### Service - AI Client (`service/ai_client.go`)

**⚠️ LEGACY - Used by JobService only**

**Purpose**: AI client interface for OpenAI integration. Used by legacy `JobService`.

**Interface**:
```go
type AIClient interface {
    AnalyzeJob(ctx, title, company, description) (*AIAnalysisResult, error)
    ResearchCompany(ctx, company, title, description) (map[string]interface{}, error)
    MatchJobToUser(ctx, job, userPrompt) (*UserMatchResult, error)
}
```

**Implementation**: `openAIClient` - Calls OpenAI GPT-3.5-turbo API

**⚠️ Note**: The new pipeline uses separate AI clients:
- `job_analysis/service/ai_client.go` - Company research only
- `user_analysis/service/ai_client.go` - User matching only

**Recommendation**: This file can be removed once `JobService` is simplified.

---

### Service - SQS Client (`service/sqs_client.go`)

**⚠️ LEGACY - Uses old queue**

**Purpose**: SQS client for sending notifications. Used by legacy `JobService`.

**Interface**:
```go
type SQSClient interface {
    SendNotification(ctx, notification) error
}
```

**Environment Variable**: `NOTIFY_SQS_QUEUE_URL` (old, not used in new pipeline)

**⚠️ Issues**:
- Uses old queue name
- Not used in production pipeline
- Only used by `JobService.queueNotification`

**Recommendation**: Remove or update to use `NOTIFICATION_QUEUE_URL`.

---

### Handler - HTTP (`handler/http.go`)

**Purpose**: HTTP handlers for job endpoints.

**Endpoints**:
- `GET /api/jobs` - Returns processed jobs (calls `JobService.GetJobs`)
- `POST /api/jobs/process` - Process a single job (local dev only, calls `JobService.ProcessJob`)

**Methods**:
- `GetJobs(w, r)` - Fetches and returns jobs
- `ProcessJob(w, r)` - Accepts job JSON, processes it, returns result

**Usage**: Used by `jobs_api` Lambda and local development server.

---

### Handler - Mock (`handler/mock.go`)

**Purpose**: Mock job fetching for local development.

**Endpoint**: `POST /api/jobs/fetch` (local dev only)

**Purpose**: Simulates the Python `jobspy_fetcher` Lambda by creating 3 mock jobs and processing them via `JobService.ProcessJob`.

**Usage**: Allows testing the full pipeline locally without running the Python worker.

---

### Handler - DTO (`handler/dto.go`)

**Purpose**: Data Transfer Objects for HTTP responses.

**Types**:
- `JobResponse` - Single job response model
- `JobsResponse` - List of jobs response model
- `ToJobResponse(job)` - Converts domain model to DTO
- `ToJobsResponse(jobs)` - Converts slice of domain models to DTO

**Usage**: Used by HTTP handlers to format responses.

---

### Handler - SQS (`handler/sqs.go`)

**⚠️ OLD - NOT USED - SHOULD BE REMOVED**

**Purpose**: Legacy SQS handler for processing jobs from SQS.

**Status**: Replaced by `job_analysis` feature's SQS handler.

**Recommendation**: **DELETE THIS FILE**.

---

### Handler - Job Analysis (`handler/job_analysis.go`)

**⚠️ OLD - NOT USED - SHOULD BE REMOVED**

**Purpose**: Legacy job analysis handler.

**Status**: Replaced by `job_analysis` feature's SQS handler.

**Recommendation**: **DELETE THIS FILE**.

---

### Module (`module.go`)

**Purpose**: Registers HTTP routes for the job feature.

**Routes**:
- `GET /jobs` - Always available
- `POST /jobs/fetch` - Only in non-production (mock jobs)
- `POST /jobs/process` - Only in non-production (process single job)

**Usage**: Called by router setup in `internal/server/router.go`.

---

### Errors (`joberr/errors.go`)

**Purpose**: Job-specific error definitions.

**Usage**: Used by repository and service layers for error handling.

---

## Issues to Fix

1. **Delete unused handlers**:
   - `handler/sqs.go` - Not used, replaced by `job_analysis/handler/sqs.go`
   - `handler/job_analysis.go` - Not used, replaced by `job_analysis/handler/sqs.go`

2. **Legacy service complexity**:
   - `JobService.ProcessJob` does too much (company research, AI analysis, user matching, notifications)
   - Should be simplified or split for local dev

3. **Legacy SQS client**:
   - `service/sqs_client.go` uses old `NOTIFY_SQS_QUEUE_URL`
   - Should be updated or removed

4. **AI client duplication**:
   - `service/ai_client.go` duplicates functionality now in `job_analysis` and `user_analysis` features
   - Can be removed once `JobService` is simplified

---

## Usage in Production

In production, the `job` feature is used **only for CRUD operations**:

1. **Job Creation**: Python `jobspy_fetcher` creates jobs directly via repository
2. **Job Reading**: `jobs_api` Lambda uses `JobService.GetJobs` to return jobs to frontend
3. **Job Processing**: Handled by 4-stage SQS pipeline (not `JobService.ProcessJob`)

---

## Usage in Local Development

For local testing, the `job` feature provides:

1. **Mock Job Fetching**: `POST /api/jobs/fetch` creates mock jobs
2. **Direct Processing**: `POST /api/jobs/process` processes a single job end-to-end
3. **Job Display**: `GET /api/jobs` shows processed jobs

This allows testing without running the full SQS pipeline.


