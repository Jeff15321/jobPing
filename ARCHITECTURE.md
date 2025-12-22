# System Architecture

## High-Level Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER INTERFACE                          │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    React Frontend                         │  │
│  │                  (Vercel / localhost:5173)               │  │
│  │                                                           │  │
│  │  • Job List Display                                      │  │
│  │  • User Preferences Form                                 │  │
│  │  • Real-time Updates                                     │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ HTTPS
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         API LAYER                               │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              REST API (Go)                                │  │
│  │         (AWS Lambda / localhost:8080)                     │  │
│  │                                                           │  │
│  │  GET  /api/v1/jobs           → List jobs                 │  │
│  │  GET  /api/v1/jobs/:id       → Get job details          │  │
│  │  POST /api/v1/users          → Create user              │  │
│  │  PUT  /api/v1/users/:id/prefs → Update preferences      │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      DATA LAYER                                 │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              PostgreSQL Database                          │  │
│  │              (AWS RDS / localhost:5432)                   │  │
│  │                                                           │  │
│  │  Tables:                                                  │  │
│  │  • jobs              → Job listings                       │  │
│  │  • users             → User accounts                      │  │
│  │  • user_job_matches  → Match results                      │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │
                              │
┌─────────────────────────────────────────────────────────────────┐
│                    BACKGROUND SERVICES                          │
│                                                                 │
│  ┌────────────────────────┐  ┌────────────────────────────┐   │
│  │   Scanner Service      │  │    Matcher Service         │   │
│  │   (Go - Cron/Lambda)   │  │    (Go - SQS Worker)       │   │
│  │                        │  │                            │   │
│  │  Every 10 minutes:     │  │  On new job:               │   │
│  │  1. Fetch from API     │  │  1. Run AI analysis        │   │
│  │  2. Store in DB        │  │  2. Match with users       │   │
│  │  3. Push to queue      │  │  3. Send notifications     │   │
│  └────────────────────────┘  └────────────────────────────┘   │
│              │                            ▲                     │
│              │                            │                     │
│              ▼                            │                     │
│  ┌────────────────────────────────────────────────────────┐   │
│  │                    SQS Queue                            │   │
│  │              (AWS SQS / LocalStack)                     │   │
│  │                                                         │   │
│  │  • Job processing queue                                 │   │
│  │  • Decouples scanner from matcher                       │   │
│  └────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   EXTERNAL INTEGRATIONS                         │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │ SpeedyApply  │  │   OpenAI     │  │      AWS SES         │ │
│  │     API      │  │     API      │  │  (Email Service)     │ │
│  │              │  │              │  │                      │ │
│  │ • Job data   │  │ • Embeddings │  │ • Send alerts        │ │
│  │ • Updates    │  │ • Analysis   │  │ • Templates          │ │
│  └──────────────┘  └──────────────┘  └──────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow

### 1. Job Fetching Flow

```
┌─────────┐     ┌──────────────┐     ┌──────────┐     ┌─────────┐
│ Scanner │────▶│ SpeedyApply  │────▶│   Jobs   │────▶│   SQS   │
│ (Cron)  │     │     API      │     │ Database │     │  Queue  │
└─────────┘     └──────────────┘     └──────────┘     └─────────┘
   Every              Fetch              Store           Notify
  10 mins            Latest              New             Matcher
                     Jobs                Jobs
```

### 2. Job Matching Flow

```
┌─────────┐     ┌──────────┐     ┌──────────┐     ┌─────────┐
│   SQS   │────▶│ Matcher  │────▶│ OpenAI   │────▶│  Email  │
│  Queue  │     │ Service  │     │   API    │     │   SES   │
└─────────┘     └──────────┘     └──────────┘     └─────────┘
   New            Process          Generate         Send
   Job            Job              Analysis         Alert
   Event                           & Match          to User
```

### 3. User Interaction Flow

```
┌─────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐
│  User   │────▶│ React   │────▶│   API    │────▶│ Database │
│ Browser │     │   UI    │     │  Server  │     │          │
└─────────┘     └─────────┘     └──────────┘     └──────────┘
   View           Display         Fetch            Query
   Jobs           Jobs            Data             Jobs
```

## Component Details

### Frontend (React + TypeScript)

**Location**: `frontend/src/`

```
Components:
├── App.tsx              → Main application
├── components/
│   ├── JobList.tsx     → Grid of job cards
│   └── JobCard.tsx     → Individual job display
├── services/
│   └── jobService.ts   → API client
└── types/
    └── job.ts          → TypeScript interfaces
```

**Responsibilities**:
- Display jobs in beautiful UI
- Handle user interactions
- Manage preferences
- Real-time updates

### Backend API (Go)

**Location**: `backend/cmd/api/`

```
Structure:
├── main.go                    → Server entry point
├── internal/api/
│   ├── handlers/
│   │   ├── jobs.go           → Job endpoints
│   │   └── users.go          → User endpoints
│   └── middleware/
│       └── middleware.go     → Logging, CORS, etc.
```

**Responsibilities**:
- Serve REST API
- Handle authentication
- Validate requests
- Return JSON responses

### Scanner Service (Go)

**Location**: `backend/cmd/scanner/`

```
Flow:
1. Wake up (every 10 minutes)
2. Call SpeedyApply API
3. Parse job data
4. Store in PostgreSQL
5. Push to SQS queue
6. Sleep
```

**Responsibilities**:
- Fetch latest jobs
- Deduplicate jobs
- Store in database
- Trigger processing

### Matcher Service (Go)

**Location**: `backend/cmd/matcher/`

```
Flow:
1. Listen to SQS queue
2. Receive new job event
3. Run AI web search
4. Generate embeddings
5. Match with users
6. Send email alerts
```

**Responsibilities**:
- AI analysis
- Semantic matching
- Email notifications
- Queue processing

### Database (PostgreSQL)

**Schema**:

```sql
jobs
├── id (PK)
├── title
├── company
├── location
├── description
├── url
├── posted_at
├── ai_analysis (JSONB)
└── created_at

users
├── id (PK)
├── email
├── preferences (TEXT)
└── created_at

user_job_matches
├── id (PK)
├── user_id (FK)
├── job_id (FK)
├── match_score
├── notified
└── created_at
```

## Deployment Architecture

### Local Development

```
┌──────────────────────────────────────────┐
│         Docker Compose                   │
│                                          │
│  ┌──────────┐  ┌──────────┐            │
│  │PostgreSQL│  │LocalStack│            │
│  │  :5432   │  │  :4566   │            │
│  └──────────┘  └──────────┘            │
└──────────────────────────────────────────┘
         ▲              ▲
         │              │
    ┌────┴────┐    ┌────┴────┐
    │   API   │    │ Scanner │
    │  :8080  │    │  (CLI)  │
    └─────────┘    └─────────┘
         ▲
         │
    ┌────┴────┐
    │  React  │
    │  :5173  │
    └─────────┘
```

### AWS Production

```
┌─────────────────────────────────────────────────┐
│                  AWS Cloud                      │
│                                                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐    │
│  │ Lambda   │  │ Lambda   │  │ Lambda   │    │
│  │  (API)   │  │(Scanner) │  │(Matcher) │    │
│  └──────────┘  └──────────┘  └──────────┘    │
│       │             │              │           │
│       │             │              │           │
│  ┌────▼─────────────▼──────────────▼────┐    │
│  │         RDS PostgreSQL                │    │
│  └───────────────────────────────────────┘    │
│                                                 │
│  ┌──────────┐  ┌──────────┐                   │
│  │   SQS    │  │   SES    │                   │
│  │  Queue   │  │  Email   │                   │
│  └──────────┘  └──────────┘                   │
└─────────────────────────────────────────────────┘
         ▲
         │
┌────────┴────────┐
│     Vercel      │
│  React Frontend │
└─────────────────┘
```

## Technology Choices

### Why Go?
- ✅ Fast compilation and execution
- ✅ Excellent concurrency (goroutines)
- ✅ Small binary size for Lambda
- ✅ Strong typing and error handling
- ✅ Great for microservices

### Why React?
- ✅ Component-based architecture
- ✅ Large ecosystem
- ✅ TypeScript support
- ✅ Fast with Vite
- ✅ Easy Vercel deployment

### Why PostgreSQL?
- ✅ JSONB for flexible AI analysis storage
- ✅ Full-text search capabilities
- ✅ ACID compliance
- ✅ Excellent performance
- ✅ Managed by AWS RDS

### Why AWS Lambda?
- ✅ Pay per execution
- ✅ Auto-scaling
- ✅ No server management
- ✅ Easy integration with SQS/SES
- ✅ Cost-effective for periodic tasks

### Why SQS?
- ✅ Decouples services
- ✅ Handles traffic spikes
- ✅ Automatic retries
- ✅ Dead letter queue
- ✅ Serverless

## Scalability Considerations

### Current Capacity
- **Jobs**: Unlimited (PostgreSQL)
- **Users**: Thousands (with current setup)
- **Requests**: ~1000/min (Lambda auto-scales)

### Scaling Strategies

**Horizontal Scaling**:
- Add more Lambda instances (automatic)
- Add read replicas for database
- Use CloudFront for frontend

**Vertical Scaling**:
- Increase Lambda memory
- Upgrade RDS instance
- Optimize queries

**Caching**:
- Redis for job data
- CloudFront for static assets
- API response caching

## Security

### Current Measures
- ✅ Environment variables for secrets
- ✅ CORS configuration
- ✅ SQL injection prevention (parameterized queries)
- ✅ HTTPS only in production

### Future Enhancements
- [ ] JWT authentication
- [ ] Rate limiting
- [ ] Input validation
- [ ] API key rotation
- [ ] Encryption at rest

## Monitoring

### Metrics to Track
- Job fetch success rate
- API response times
- Database query performance
- Lambda execution duration
- Email delivery rate
- Match accuracy

### Tools
- CloudWatch Logs
- CloudWatch Metrics
- CloudWatch Alarms
- RDS Performance Insights

## Cost Estimation

### Monthly AWS Costs (Production)

```
RDS db.t3.micro:        $15
Lambda executions:      $2-5
SQS messages:           $0 (free tier)
SES emails:             $0-1
Data transfer:          $1-2
────────────────────────────
Total:                  ~$20-25/month
```

### Vercel
- Free tier for frontend
- $0/month

**Total Monthly Cost**: ~$20-25

## Performance Targets

- **API Response Time**: < 200ms
- **Job Fetch Time**: < 5s
- **Matching Time**: < 10s per job
- **Email Delivery**: < 30s
- **Frontend Load**: < 2s

---

For implementation details, see:
- [SETUP.md](SETUP.md) - Development setup
- [PROJECT_STATUS.md](PROJECT_STATUS.md) - Current status
- [QUICKSTART.md](QUICKSTART.md) - Quick start guide
