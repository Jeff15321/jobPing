# JobPing Architecture Plan

## Table of Contents

- [System Overview](#system-overview)
- [Database Schema](#database-schema)
- [File Structure](#file-structure)
- [API Endpoints](#api-endpoints)
- [SQS Message Formats](#sqs-message-formats)
- [Pre-Filter JSON Schema](#pre-filter-json-schema)
- [LLM Scoring Prompt](#llm-scoring-prompt)
- [Deduplication Logic](#deduplication-logic)
- [Processing Flow](#processing-flow)
- [Terraform Resources](#terraform-resources)
- [Configuration Variables](#configuration-variables)
- [Retention Cleanup](#retention-cleanup)
- [Summary](#summary)

---

## System Overview

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                    AWS Cloud                                         │
│                                                                                      │
│  ┌──────────────┐     ┌─────────────────┐     ┌─────────────────┐                   │
│  │  EventBridge │────▶│  ECS: JobSpy    │────▶│  SQS: New Jobs  │                   │
│  │  (Cron 1hr)  │     │  (Python)       │     │                 │                   │
│  └──────────────┘     └─────────────────┘     └────────┬────────┘                   │
│                              │                         │                             │
│                              │ New Company?            │                             │
│                              ▼                         │                             │
│                       ┌─────────────────┐              │                             │
│                       │ Perplexity API  │              │                             │
│                       │ (Company Research)             │                             │
│                       └─────────────────┘              │                             │
│                              │                         │                             │
│                              ▼                         ▼                             │
│  ┌─────────────────────────────────────────────────────────────────┐                │
│  │                      Go Backend (ECS/Lambda)                     │                │
│  │  ┌─────────────┐  ┌──────────────┐  ┌────────────────────────┐  │                │
│  │  │ User API    │  │ SQS Consumer │  │ LLM Scoring Service    │  │                │
│  │  │ Preferences │  │ Pre-filter   │  │ (GPT-4 Mini)           │  │                │
│  │  │ Job History │  │              │  │                        │  │                │
│  │  └─────────────┘  └──────────────┘  └────────────────────────┘  │                │
│  └─────────────────────────────────────────────────────────────────┘                │
│                              │                                                       │
│                              │ Match found                                           │
│                              ▼                                                       │
│                       ┌─────────────────┐     ┌─────────────────┐                   │
│                       │ SQS: Notify     │────▶│ ECS: Apprise    │────▶ Email        │
│                       │                 │     │ (Python)        │                   │
│                       └─────────────────┘     └─────────────────┘                   │
│                                                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐                │
│  │                         RDS (PostgreSQL)                         │                │
│  │  users, preferences, jobs, companies, job_matches, job_rejects  │                │
│  └─────────────────────────────────────────────────────────────────┘                │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              Vercel (Frontend)                                       │
│  ┌─────────────────────────────────────────────────────────────────┐                │
│  │  React App: Preferences UI, Matched Jobs, Rejected Jobs         │                │
│  └─────────────────────────────────────────────────────────────────┘                │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### Key Technologies

| Component | Technology | Source |
|-----------|------------|--------|
| Job Scraping | Python + jobspy-api | [rainmanjam/jobspy-api](https://github.com/rainmanjam/jobspy-api) |
| Notifications | Python + Apprise | [caronc/apprise](https://github.com/caronc/apprise) |
| Backend API & LLM Scoring | Go | Existing codebase |
| Message Queue | AWS SQS | - |
| Scheduler | AWS EventBridge | - |
| Database | PostgreSQL | AWS RDS |
| Container Orchestration | AWS ECS | - |

---

## Database Schema

### Tables

```sql
-- ============================================
-- USERS (existing, extend if needed)
-- ============================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    notification_email VARCHAR(255), -- can differ from login email
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ============================================
-- USER PREFERENCES (max 3 per user)
-- ============================================
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,                    -- e.g., "Backend Roles in Toronto"
    raw_prompt TEXT NOT NULL,                       -- user's natural language prompt
    filter_json JSONB NOT NULL,                     -- LLM-generated pre-filter JSON
    match_threshold INTEGER NOT NULL DEFAULT 70,   -- 0-100, user configurable
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT max_preferences_per_user CHECK (
        (SELECT COUNT(*) FROM user_preferences WHERE user_id = user_preferences.user_id) <= 3
    )
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
CREATE INDEX idx_user_preferences_active ON user_preferences(is_active) WHERE is_active = TRUE;

-- ============================================
-- COMPANIES (cached research data)
-- ============================================
CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    normalized_name VARCHAR(255) NOT NULL UNIQUE,  -- lowercase, trimmed for dedup
    raw_research TEXT,                              -- raw Perplexity response
    summary_json JSONB,                             -- summarized company info
    research_source VARCHAR(50) DEFAULT 'perplexity',
    researched_at TIMESTAMP,
    expires_at TIMESTAMP,                           -- 6 months from researched_at
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_companies_normalized_name ON companies(normalized_name);
CREATE INDEX idx_companies_expires_at ON companies(expires_at);

-- ============================================
-- JOBS (scraped and deduplicated)
-- ============================================
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fingerprint VARCHAR(64) UNIQUE NOT NULL,       -- SHA256(normalized company + title + location)
    
    -- Raw job data
    title VARCHAR(500) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    company_id UUID REFERENCES companies(id),
    location VARCHAR(255),
    job_type VARCHAR(50),                          -- fulltime, parttime, contract, internship
    is_remote BOOLEAN,
    salary_min INTEGER,
    salary_max INTEGER,
    salary_currency VARCHAR(10),
    description TEXT,
    url VARCHAR(2000) NOT NULL,
    source VARCHAR(50) NOT NULL,                   -- linkedin, indeed, glassdoor, etc.
    
    -- Metadata
    posted_at TIMESTAMP,
    scraped_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_jobs_fingerprint ON jobs(fingerprint);
CREATE INDEX idx_jobs_company_id ON jobs(company_id);
CREATE INDEX idx_jobs_scraped_at ON jobs(scraped_at);
CREATE INDEX idx_jobs_source ON jobs(source);

-- ============================================
-- JOB MATCHES (accepted jobs per user preference)
-- Retention: 3 months (configurable)
-- ============================================
CREATE TABLE job_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    preference_id UUID NOT NULL REFERENCES user_preferences(id) ON DELETE CASCADE,
    
    score INTEGER NOT NULL,                        -- 0-100 from LLM
    match_reason TEXT NOT NULL,                    -- LLM explanation of why matched
    
    notified_at TIMESTAMP,                         -- when email was sent
    viewed_at TIMESTAMP,                           -- when user viewed in UI
    
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,                 -- 3 months from created_at
    
    UNIQUE(job_id, preference_id)
);

CREATE INDEX idx_job_matches_user_id ON job_matches(user_id);
CREATE INDEX idx_job_matches_preference_id ON job_matches(preference_id);
CREATE INDEX idx_job_matches_expires_at ON job_matches(expires_at);
CREATE INDEX idx_job_matches_created_at ON job_matches(created_at DESC);

-- ============================================
-- JOB REJECTS (jobs that passed pre-filter but failed threshold)
-- Retention: 1 week (configurable)
-- ============================================
CREATE TABLE job_rejects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    preference_id UUID NOT NULL REFERENCES user_preferences(id) ON DELETE CASCADE,
    
    score INTEGER NOT NULL,                        -- 0-100 from LLM
    reject_reason TEXT NOT NULL,                   -- brief LLM explanation
    
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,                 -- 1 week from created_at
    
    UNIQUE(job_id, preference_id)
);

CREATE INDEX idx_job_rejects_user_id ON job_rejects(user_id);
CREATE INDEX idx_job_rejects_preference_id ON job_rejects(preference_id);
CREATE INDEX idx_job_rejects_expires_at ON job_rejects(expires_at);

-- ============================================
-- SYSTEM CONFIG (for adjustable values)
-- ============================================
CREATE TABLE system_config (
    key VARCHAR(100) PRIMARY KEY,
    value VARCHAR(500) NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert default configs
INSERT INTO system_config (key, value, description) VALUES
    ('max_preferences_per_user', '3', 'Maximum number of preference profiles per user'),
    ('scrape_interval_minutes', '60', 'How often to run job scraping cron'),
    ('company_cache_days', '180', 'How long to cache company research (6 months)'),
    ('match_retention_days', '90', 'How long to keep matched jobs (3 months)'),
    ('reject_retention_days', '7', 'How long to keep rejected jobs (1 week)'),
    ('default_match_threshold', '70', 'Default match score threshold (0-100)');
```

---

## File Structure

```
jobPing/
├── backend/                              # Go backend (existing)
│   ├── cmd/
│   │   ├── server/
│   │   │   └── main.go
│   │   ├── lambda/
│   │   │   └── main.go
│   │   └── worker/                       # NEW: SQS consumer worker
│   │       └── main.go
│   ├── internal/
│   │   ├── app/
│   │   │   └── app.go
│   │   ├── config/
│   │   │   └── config.go                 # Add new config vars
│   │   ├── database/
│   │   │   ├── db.go
│   │   │   ├── migrate.go
│   │   │   └── migrations/
│   │   │       ├── 000001_init_schema.down.sql
│   │   │       ├── 000001_init_schema.up.sql
│   │   │       ├── 000002_job_matching.up.sql    # NEW
│   │   │       └── 000002_job_matching.down.sql  # NEW
│   │   ├── features/
│   │   │   ├── user/                     # Existing
│   │   │   │   ├── handler/
│   │   │   │   ├── model/
│   │   │   │   ├── repository/
│   │   │   │   ├── service/
│   │   │   │   └── module.go
│   │   │   ├── preference/               # NEW
│   │   │   │   ├── handler/
│   │   │   │   │   ├── dto.go
│   │   │   │   │   └── http.go
│   │   │   │   ├── model/
│   │   │   │   │   └── preference.go
│   │   │   │   ├── repository/
│   │   │   │   │   └── preference_repository.go
│   │   │   │   ├── service/
│   │   │   │   │   └── preference_service.go
│   │   │   │   └── module.go
│   │   │   ├── job/                      # NEW
│   │   │   │   ├── handler/
│   │   │   │   │   ├── dto.go
│   │   │   │   │   └── http.go
│   │   │   │   ├── model/
│   │   │   │   │   ├── job.go
│   │   │   │   │   ├── job_match.go
│   │   │   │   │   └── job_reject.go
│   │   │   │   ├── repository/
│   │   │   │   │   ├── job_repository.go
│   │   │   │   │   ├── match_repository.go
│   │   │   │   │   └── reject_repository.go
│   │   │   │   ├── service/
│   │   │   │   │   └── job_service.go
│   │   │   │   └── module.go
│   │   │   └── company/                  # NEW
│   │   │       ├── model/
│   │   │       │   └── company.go
│   │   │       ├── repository/
│   │   │       │   └── company_repository.go
│   │   │       ├── service/
│   │   │       │   └── company_service.go
│   │   │       └── module.go
│   │   ├── llm/                          # NEW: LLM integration
│   │   │   ├── client.go                 # Interface for LLM providers
│   │   │   ├── openai/
│   │   │   │   └── client.go             # OpenAI/GPT-4 Mini implementation
│   │   │   ├── prompts/
│   │   │   │   ├── filter_extraction.go  # Prompt for extracting filter JSON
│   │   │   │   ├── job_scoring.go        # Prompt for scoring jobs
│   │   │   │   └── company_summary.go    # Prompt for summarizing company research
│   │   │   └── types.go
│   │   ├── queue/                        # NEW: SQS integration
│   │   │   ├── consumer.go
│   │   │   ├── producer.go
│   │   │   └── messages.go               # Message type definitions
│   │   ├── matching/                     # NEW: Job matching logic
│   │   │   ├── prefilter.go              # Pre-filter logic
│   │   │   ├── scorer.go                 # LLM scoring wrapper
│   │   │   └── processor.go              # Main processing orchestrator
│   │   └── server/
│   │       └── router.go                 # Add new routes
│   └── go.mod
│
├── services/                             # NEW: Python services
│   ├── job-scraper/                      # Job scraping service
│   │   ├── Dockerfile
│   │   ├── requirements.txt
│   │   ├── src/
│   │   │   ├── __init__.py
│   │   │   ├── main.py                   # Entry point
│   │   │   ├── scraper.py                # JobSpy wrapper
│   │   │   ├── deduplicator.py           # Fingerprinting logic
│   │   │   ├── company_researcher.py     # Perplexity API integration
│   │   │   ├── queue_publisher.py        # SQS publisher
│   │   │   └── config.py
│   │   └── tests/
│   └── apprise-notifier/                 # Notification service
│       ├── Dockerfile
│       ├── requirements.txt
│       ├── src/
│       │   ├── __init__.py
│       │   ├── main.py                   # SQS consumer + Apprise sender
│       │   ├── notifier.py               # Apprise wrapper
│       │   ├── templates/
│       │   │   └── job_match_email.html
│       │   └── config.py
│       └── tests/
│
├── frontend/                             # Existing React app
│   ├── src/
│   │   ├── components/
│   │   │   ├── JobCard.tsx               # Existing
│   │   │   ├── JobList.tsx               # Existing
│   │   │   ├── PreferenceCard.tsx        # NEW
│   │   │   ├── PreferenceForm.tsx        # NEW
│   │   │   └── MatchReasonBadge.tsx      # NEW
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx             # NEW: Main dashboard
│   │   │   ├── Preferences.tsx           # NEW: Manage preferences
│   │   │   ├── MatchedJobs.tsx           # NEW: View matched jobs
│   │   │   └── RejectedJobs.tsx          # NEW: View rejected jobs
│   │   ├── services/
│   │   │   ├── api.ts
│   │   │   ├── jobService.ts
│   │   │   └── preferenceService.ts      # NEW
│   │   └── types/
│   │       ├── job.ts
│   │       └── preference.ts             # NEW
│   └── ...
│
├── infra/
│   └── terraform/
│       ├── main.tf
│       ├── lambda.tf                     # Existing
│       ├── ecs.tf                        # NEW: ECS cluster + services
│       ├── sqs.tf                        # NEW: SQS queues
│       ├── eventbridge.tf                # NEW: Cron scheduler
│       ├── ecr.tf                        # NEW: Container registries
│       └── iam.tf                        # NEW: IAM roles for services
│
└── docs/
    └── ARCHITECTURE.md                   # This file
```

---

## API Endpoints

### Preferences API

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/preferences` | List user's preferences (max 3) |
| `GET` | `/api/v1/preferences/:id` | Get single preference with filter JSON |
| `POST` | `/api/v1/preferences` | Create new preference (triggers LLM filter extraction) |
| `PUT` | `/api/v1/preferences/:id` | Update preference (re-triggers LLM) |
| `DELETE` | `/api/v1/preferences/:id` | Delete preference |
| `PATCH` | `/api/v1/preferences/:id/threshold` | Update just the threshold |
| `POST` | `/api/v1/preferences/:id/toggle` | Enable/disable preference |

### Jobs API

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/jobs/matched` | List matched jobs (3 months, paginated) |
| `GET` | `/api/v1/jobs/matched/:id` | Get matched job with full reason |
| `GET` | `/api/v1/jobs/rejected` | List rejected jobs (1 week, paginated) |
| `GET` | `/api/v1/jobs/rejected/:id` | Get rejected job with reason |
| `POST` | `/api/v1/jobs/:id/viewed` | Mark job as viewed |

### Query Parameters for Job Lists

```
?preference_id=<uuid>     # Filter by specific preference
?page=1&limit=20          # Pagination
?sort=score|date          # Sort order
?order=asc|desc           # Sort direction
```

---

## SQS Message Formats

### Queue 1: `jobping-new-jobs`

**Published by:** Job Scraper (Python)  
**Consumed by:** Go Backend Worker

```json
{
  "message_type": "new_job",
  "job": {
    "fingerprint": "sha256hash...",
    "title": "Senior Software Engineer",
    "company_name": "Stripe",
    "company_normalized": "stripe",
    "location": "Toronto, ON, Canada",
    "job_type": "fulltime",
    "is_remote": false,
    "salary_min": 150000,
    "salary_max": 200000,
    "salary_currency": "CAD",
    "description": "We are looking for...",
    "url": "https://linkedin.com/jobs/...",
    "source": "linkedin",
    "posted_at": "2025-12-28T10:00:00Z"
  },
  "company": {
    "name": "Stripe",
    "normalized_name": "stripe",
    "is_new": true,
    "research": {
      "raw": "Perplexity response...",
      "summary": {
        "prestige_score": 9,
        "employee_satisfaction": 8.5,
        "benefits": ["unlimited PTO", "health insurance", "free lunch"],
        "work_life_balance": 7,
        "growth_opportunities": 9,
        "tech_stack": ["Go", "Ruby", "React"],
        "company_size": "1000-5000",
        "industry": "Fintech",
        "funding_stage": "public",
        "notable_info": "Known for excellent engineering culture"
      }
    }
  },
  "scraped_at": "2025-12-28T12:00:00Z"
}
```

### Queue 2: `jobping-notifications`

**Published by:** Go Backend (after match found)  
**Consumed by:** Apprise Notifier (Python)

```json
{
  "message_type": "job_match_notification",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "notification_email": "user@example.com"
  },
  "preference": {
    "id": "uuid",
    "name": "Backend Roles in Toronto"
  },
  "job": {
    "id": "uuid",
    "title": "Senior Software Engineer",
    "company_name": "Stripe",
    "location": "Toronto, ON, Canada",
    "url": "https://linkedin.com/jobs/...",
    "salary_range": "$150,000 - $200,000 CAD"
  },
  "match": {
    "score": 87,
    "reason": "Strong match: Stripe is a highly prestigious fintech company known for excellent engineering culture and benefits. The role is backend-focused in Toronto, matching your location preference. Company offers free lunch and unlimited PTO which aligns with your benefits requirements."
  }
}
```

---

## Pre-Filter JSON Schema

When a user creates a preference with a natural language prompt, the LLM extracts a structured filter JSON.

### IMPORTANT: Only Non-Null Fields

**The LLM should ONLY include fields that the user explicitly mentioned or strongly implied.** Do NOT include fields with null, empty, or default values. This saves space and makes pre-filtering more efficient.

For example, if the user doesn't mention salary requirements, do NOT include `min_salary` or `max_salary` in the JSON.

### LLM Prompt for Filter Extraction

```
You are a job preference analyzer. Given a user's natural language job search requirements, 
extract a structured JSON filter that can be used to pre-filter jobs before detailed analysis.

IMPORTANT RULES:
1. Only include fields that the user explicitly mentioned or strongly implied
2. Do NOT include fields with null, empty, or default values
3. For array fields, only include if user specified requirements
4. Be strict - if user didn't mention something, don't assume it

AVAILABLE FIELDS (include only if relevant):

{
  "locations_include": ["array of locations user wants"],
  "locations_exclude": ["array of locations user does NOT want"],
  "keywords_include": ["job title keywords user wants"],
  "keywords_exclude": ["keywords user explicitly does NOT want"],
  "company_keywords_include": ["company name patterns to include"],
  "company_keywords_exclude": ["company name patterns to exclude"],
  "job_types": ["fulltime", "parttime", "contract", "internship"],
  "min_salary": 100000,
  "max_salary": 200000,
  "salary_currency": "USD",
  "experience_levels": ["entry", "mid", "senior", "lead", "executive"],
  "company_sizes": ["startup", "small", "medium", "large", "enterprise"],
  "industries": ["fintech", "healthcare", "e-commerce", ...],
  "must_have_salary_posted": true,
  "posted_within_hours": 24
}

USER'S PROMPT:
"{user_prompt}"

OUTPUT: Return ONLY the JSON object with relevant fields. No explanations.
```

### Example Extraction

**User Prompt:**
> "I want well known companies only that has a good reputation with employee benefits. Preferably in Toronto, Canada, but it is fine if it is in the US if it's in the bay area or new york, or if the listing is very valuable. Only look for Software Engineering Roles, backend or full stack also works, but nothing related to hardware or UI/UX."

**Extracted Filter JSON:**

```json
{
  "locations_include": ["Toronto", "Canada", "Bay Area", "San Francisco", "New York", "NYC"],
  "keywords_include": ["Software Engineer", "Backend Engineer", "Full Stack Engineer", "Software Developer", "Backend Developer", "Full Stack Developer"],
  "keywords_exclude": ["Hardware", "UI/UX", "UX Designer", "UI Designer", "Frontend", "Embedded", "Firmware"],
  "job_types": ["fulltime"]
}
```

**Note:** Fields like `company_sizes`, `industries`, `min_salary` are NOT included because the user didn't specify them. The LLM scoring step will still consider "well known companies" and "good reputation with employee benefits" but these aren't pre-filterable from job listings alone.

---

## LLM Scoring Prompt

```
You are a job matching assistant. Evaluate how well a job posting matches a user's requirements.

USER'S REQUIREMENTS (natural language):
"{user_raw_prompt}"

JOB POSTING:
- Title: {job.title}
- Company: {job.company_name}
- Location: {job.location}
- Type: {job.job_type}
- Remote: {job.is_remote}
- Salary: {job.salary_min} - {job.salary_max} {job.salary_currency}
- Description: {job.description (truncated to 2000 chars)}

COMPANY RESEARCH:
{company.summary_json}

SCORING CRITERIA:
1. Location match (does it meet user's geographic requirements?)
2. Role match (does the job title/description match what user wants?)
3. Company quality (does the company meet user's prestige/culture requirements?)
4. Benefits alignment (does the company offer what user values?)
5. Any explicit exclusions violated?

OUTPUT FORMAT (JSON only):
{
  "score": <0-100>,
  "reason": "<2-3 sentences explaining the score>",
  "highlights": ["<key positive match point 1>", "<key positive match point 2>"],
  "concerns": ["<any concern or partial mismatch>"]
}

If the job clearly violates an explicit exclusion (e.g., user said no hardware roles and this is a hardware role), score should be 0-20.
```

---

## Deduplication Logic

Located in: `services/job-scraper/src/deduplicator.py`

The deduplication strategy uses a fingerprint hash based on normalized company name, job title, and location. This ensures the same job appearing on multiple boards (LinkedIn, Indeed, etc.) is only processed once.

### Algorithm

```python
import hashlib
import re

def normalize_string(s: str) -> str:
    """Normalize string for consistent fingerprinting."""
    if not s:
        return ""
    # Lowercase
    s = s.lower()
    # Remove special characters except spaces
    s = re.sub(r'[^a-z0-9\s]', '', s)
    # Collapse multiple spaces
    s = re.sub(r'\s+', ' ', s)
    # Trim
    return s.strip()

def normalize_company(company: str) -> str:
    """Normalize company name, removing common suffixes."""
    company = normalize_string(company)
    # Remove common suffixes
    suffixes = ['inc', 'llc', 'ltd', 'corp', 'corporation', 'company', 'co']
    for suffix in suffixes:
        if company.endswith(f' {suffix}'):
            company = company[:-len(suffix)-1]
    return company.strip()

def normalize_location(location: str) -> str:
    """Normalize location to city level."""
    location = normalize_string(location)
    # Common patterns to simplify
    # "San Francisco, CA, USA" -> "san francisco ca"
    # "Toronto, ON, Canada" -> "toronto on"
    parts = location.split(',')
    if len(parts) >= 2:
        # Take city and state/province only
        return ' '.join(normalize_string(p) for p in parts[:2])
    return location

def generate_fingerprint(company: str, title: str, location: str) -> str:
    """Generate unique fingerprint for job deduplication."""
    normalized = f"{normalize_company(company)}|{normalize_string(title)}|{normalize_location(location)}"
    return hashlib.sha256(normalized.encode()).hexdigest()
```

### What This Ensures

| Scenario | Result |
|----------|--------|
| Same job from LinkedIn and Indeed | Same fingerprint → deduplicated |
| "Stripe Inc." and "Stripe" | Same fingerprint |
| "San Francisco, CA, USA" and "San Francisco, California" | Same fingerprint |
| Job reposts | Same fingerprint → won't create duplicates |

---

## Processing Flow

### Step 1: Job Scraping (Every 1 Hour)

```
EventBridge Cron → ECS Job Scraper

1. Call jobspy-api for each source (LinkedIn, Indeed, Glassdoor, etc.)
2. For each job:
   a. Generate fingerprint
   b. Check if fingerprint exists in DB (call Go API or direct DB)
   c. If new job:
      - Check if company exists in DB
      - If new company → Call Perplexity API → Summarize with LLM
      - Publish to SQS: jobping-new-jobs
```

### Step 2: Job Processing (Go Worker)

```
SQS Consumer (Go) polls jobping-new-jobs

For each message:
1. Save job to DB (with company reference)
2. For each ACTIVE user preference:
   a. Run pre-filter (keywords, location, exclusions)
   b. If passes pre-filter:
      - Call LLM for scoring
      - If score >= user's threshold:
        → Save to job_matches
        → Publish to SQS: jobping-notifications
      - Else:
        → Save to job_rejects
   c. If fails pre-filter:
      → Skip (don't save to rejects, too many)
```

### Step 3: Notifications (Apprise)

```
SQS Consumer (Python) polls jobping-notifications

For each message:
1. Format email using template
2. Send via Apprise (email)
3. Acknowledge message
```

---

## Terraform Resources

### New Resources to Create

| Resource | Name | Purpose |
|----------|------|---------|
| ECS Cluster | `jobping-cluster` | Shared cluster for all services |
| ECS Service | `jobping-scraper` | Job scraper (jobspy-api based) |
| ECS Service | `jobping-notifier` | Apprise notification service |
| ECS Service | `jobping-worker` | Go SQS consumer (optional, can be Lambda) |
| ECR Repo | `jobping-scraper` | Docker images for scraper |
| ECR Repo | `jobping-notifier` | Docker images for notifier |
| SQS Queue | `jobping-new-jobs` | New job events |
| SQS Queue | `jobping-new-jobs-dlq` | Dead letter queue |
| SQS Queue | `jobping-notifications` | Notification events |
| SQS Queue | `jobping-notifications-dlq` | Dead letter queue |
| EventBridge Rule | `jobping-scrape-schedule` | Cron: every 1 hour |
| Secrets Manager | `jobping/perplexity` | Perplexity API key |
| Secrets Manager | `jobping/openai` | OpenAI API key |
| Secrets Manager | `jobping/smtp` | SMTP credentials for Apprise |

---

## Configuration Variables

### Environment Variables (Go Backend)

```env
# LLM Configuration
LLM_PROVIDER=openai                    # openai, anthropic, etc. (easily swappable)
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4o-mini

# Queue Configuration  
AWS_REGION=us-east-1
SQS_NEW_JOBS_QUEUE_URL=https://sqs...
SQS_NOTIFICATIONS_QUEUE_URL=https://sqs...

# Retention (adjustable)
MATCH_RETENTION_DAYS=90
REJECT_RETENTION_DAYS=7
COMPANY_CACHE_DAYS=180

# Limits
MAX_PREFERENCES_PER_USER=3
```

### Environment Variables (Job Scraper)

```env
# JobSpy API (internal call to jobspy-api container)
JOBSPY_API_URL=http://localhost:8000

# Perplexity
PERPLEXITY_API_KEY=pplx-...

# Queue
AWS_REGION=us-east-1
SQS_NEW_JOBS_QUEUE_URL=https://sqs...

# Database (for fingerprint checking)
DATABASE_URL=postgres://...

# Scraping config
SCRAPE_INTERVAL_MINUTES=60
RESULTS_PER_SITE=50
JOB_SITES=linkedin,indeed,glassdoor,ziprecruiter
```

### Environment Variables (Apprise Notifier)

```env
# Apprise configuration
APPRISE_URLS=mailto://user:pass@smtp.example.com

# Queue
AWS_REGION=us-east-1
SQS_NOTIFICATIONS_QUEUE_URL=https://sqs...
```

---

## Retention Cleanup

A scheduled job (can be Lambda or cron in Go worker) runs daily:

```sql
-- Delete expired matches (older than 3 months)
DELETE FROM job_matches WHERE expires_at < NOW();

-- Delete expired rejects (older than 1 week)
DELETE FROM job_rejects WHERE expires_at < NOW();

-- Delete expired company cache (older than 6 months)
UPDATE companies SET raw_research = NULL, summary_json = NULL 
WHERE expires_at < NOW();

-- Optional: Delete orphaned jobs (no matches/rejects referencing them)
-- Run weekly or monthly
DELETE FROM jobs 
WHERE id NOT IN (SELECT job_id FROM job_matches)
  AND id NOT IN (SELECT job_id FROM job_rejects)
  AND scraped_at < NOW() - INTERVAL '7 days';
```

---

## Summary

| Component | Technology | Deployment |
|-----------|------------|------------|
| User API, Preferences, Job History, LLM Scoring | Go | Lambda / ECS |
| SQS Worker (job processing) | Go | Lambda / ECS |
| Job Scraper | Python + [jobspy-api](https://github.com/rainmanjam/jobspy-api) | ECS |
| Notification Service | Python + [Apprise](https://github.com/caronc/apprise) | ECS |
| Database | PostgreSQL | RDS |
| Queues | SQS | AWS |
| Scheduler | EventBridge | AWS |
| Frontend | React | Vercel |

---

## Adjustable Parameters Summary
 
| Parameter | Default | Location |
|-----------|---------|----------|
| Max preferences per user | 3 | `system_config` table / env var |
| Scrape interval | 60 minutes | EventBridge cron / env var |
| Company cache duration | 6 months (180 days) | `system_config` table / env var |
| Match retention | 3 months (90 days) | `system_config` table / env var |
| Reject retention | 1 week (7 days) | `system_config` table / env var |
| Default match threshold | 70 | `system_config` table / env var |
| LLM model | gpt-4o-mini | env var |
| LLM provider | openai | env var |

