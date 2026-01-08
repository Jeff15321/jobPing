# User Feature Documentation

## Overview

The `user` feature provides **user management** functionality: registration, authentication, profile management, and user preferences. It handles all user-related HTTP endpoints and JWT authentication.

## File Structure

```
backend/internal/features/user/
├── handler/
│   ├── dto.go              # Data Transfer Objects (request/response models)
│   ├── http.go             # HTTP handlers for user endpoints
│   └── middleware.go       # JWT authentication middleware
├── model/
│   └── user.go             # User domain models
├── module.go               # Route registration
├── repository/
│   └── user_repository.go  # Database operations for users
├── service/
│   └── user_service.go     # Business logic for user operations
└── usererr/
    └── errors.go           # User-specific error definitions
```

## Components

### Model (`model/user.go`)

**Purpose**: Defines user-related domain models.

**Types**:

1. **User**:
```go
type User struct {
    ID              uuid.UUID
    Username        string
    PasswordHash    string
    AIPrompt        *string              // User's ideal job description
    DiscordWebhook  *string              // Discord webhook URL (optional)
    NotifyThreshold int                   // Minimum score to notify (default: 70)
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

2. **UserJobMatch**:
```go
type UserJobMatch struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    JobID     uuid.UUID
    Score     int                    // Match score (0-100)
    Analysis  map[string]interface{} // Full AI analysis
    Notified  bool                   // Whether user was notified
    CreatedAt time.Time
}
```

3. **Preference**:
```go
type Preference struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Key       string                 // Preference key (e.g., "location", "salary")
    Value     string                 // Preference value
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Usage**: Used by repository and service layers to represent user data.

---

### Repository (`repository/user_repository.go`)

**Purpose**: Database operations for users, matches, and preferences.

**Interfaces**:

1. **UserRepository**:
```go
type UserRepository interface {
    CreateUser(ctx, user) error
    GetUserByUsername(ctx, username) (*User, error)
    GetUserByID(ctx, id) (*User, error)
    UpdateAIPrompt(ctx, userID, prompt) error
    UpdateDiscordWebhook(ctx, userID, webhook) error
    UpdateNotifyThreshold(ctx, userID, threshold) error
    GetUsersWithPrompts(ctx) ([]User, error)  // Used by fanout stage
}
```

2. **UserJobMatchRepository**:
```go
type UserJobMatchRepository interface {
    Create(ctx, match) error
    GetByUserID(ctx, userID) ([]UserJobMatch, error)
    GetByUserAndJob(ctx, userID, jobID) (*UserJobMatch, error)
    MarkNotified(ctx, id) error
    GetUnnotifiedAboveThreshold(ctx) ([]UserJobMatch, error)
}
```

3. **PreferenceRepository**:
```go
type PreferenceRepository interface {
    Create(ctx, pref) error
    GetByUserID(ctx, userID) ([]Preference, error)
    GetByID(ctx, id) (*Preference, error)
    GetByUserIDAndKey(ctx, userID, key) (*Preference, error)
    Update(ctx, pref) error
    Delete(ctx, id) error
}
```

**Database Tables**:
- `users` - User accounts
- `user_job_matches` - Job matches for users
- `preferences` - User preferences (key-value pairs)

**Usage**: Used by service layer and other features (fanout, analysis, notification).

---

### Service (`service/user_service.go`)

**Purpose**: Business logic for user operations.

**Key Methods**:

1. **Register**:
```go
Register(ctx context.Context, username, password string) (*User, error)
```
- Validates username doesn't exist
- Hashes password with bcrypt
- Creates user in database
- Returns user (without password hash)

2. **Login**:
```go
Login(ctx context.Context, username, password string) (string, error)
```
- Validates username and password
- Returns JWT token on success

3. **UpdateAIPrompt**:
```go
UpdateAIPrompt(ctx context.Context, userID uuid.UUID, prompt string) error
```
- Updates user's AI prompt (ideal job description)
- Used by pipeline for job matching

4. **UpdateDiscordWebhook**:
```go
UpdateDiscordWebhook(ctx context.Context, userID uuid.UUID, webhook string) error
```
- Updates user's Discord webhook URL
- Used for notifications (future)

5. **UpdateNotifyThreshold**:
```go
UpdateNotifyThreshold(ctx context.Context, userID uuid.UUID, threshold int) error
```
- Updates minimum match score to trigger notification
- Default: 70

6. **GetUser**:
```go
GetUser(ctx context.Context, userID uuid.UUID) (*User, error)
```
- Returns user by ID (without password hash)

**Dependencies**:
- `UserRepository` - Database operations
- `PreferenceRepository` - Preference operations
- `UserJobMatchRepository` - Match operations
- JWT secret and expiry from config

**Error Handling**:
- Returns domain-specific errors (`usererr.ErrUserAlreadyExists`, etc.)

---

### Handler - HTTP (`handler/http.go`)

**Purpose**: HTTP handlers for user endpoints.

**Endpoints**:

1. **POST /api/register**:
```go
Register(w http.ResponseWriter, r *http.Request)
```
- Accepts: `{username, password}`
- Returns: `{user: {id, username, ...}}`
- Creates new user account

2. **POST /api/login**:
```go
Login(w http.ResponseWriter, r *http.Request)
```
- Accepts: `{username, password}`
- Returns: `{token: "jwt-token"}`
- Authenticates user and returns JWT

3. **GET /api/user** (protected):
```go
GetUser(w http.ResponseWriter, r *http.Request)
```
- Requires: JWT token in `Authorization` header
- Returns: `{user: {id, username, ai_prompt, ...}}`
- Returns current user's profile

4. **PUT /api/user/ai-prompt** (protected):
```go
UpdateAIPrompt(w http.ResponseWriter, r *http.Request)
```
- Requires: JWT token
- Accepts: `{ai_prompt: "..."}`
- Updates user's AI prompt

5. **PUT /api/user/discord-webhook** (protected):
```go
UpdateDiscordWebhook(w http.ResponseWriter, r *http.Request)
```
- Requires: JWT token
- Accepts: `{discord_webhook: "..."}`
- Updates user's Discord webhook

6. **PUT /api/user/notify-threshold** (protected):
```go
UpdateNotifyThreshold(w http.ResponseWriter, r *http.Request)
```
- Requires: JWT token
- Accepts: `{notify_threshold: 70}`
- Updates user's notification threshold

**Usage**: Used by `api` Lambda and local development server.

---

### Handler - Middleware (`handler/middleware.go`)

**Purpose**: JWT authentication middleware.

**Key Component**:
```go
type AuthMiddleware struct {
    jwtSecret  string
    jwtExpiry  time.Duration
}
```

**Methods**:

1. **NewAuthMiddleware**:
```go
NewAuthMiddleware(secret string, expiry time.Duration) *AuthMiddleware
```
- Creates middleware instance

2. **GenerateToken**:
```go
GenerateToken(userID uuid.UUID) (string, error)
```
- Generates JWT token for user
- Used by login handler

3. **ValidateToken**:
```go
ValidateToken(tokenString string) (uuid.UUID, error)
```
- Validates JWT token
- Returns user ID if valid

4. **Middleware**:
```go
Middleware(next http.Handler) http.Handler
```
- HTTP middleware function
- Extracts token from `Authorization: Bearer <token>` header
- Validates token and adds user ID to request context
- Returns 401 if token invalid/missing

**Usage**: Applied to protected routes via router.

---

### Handler - DTO (`handler/dto.go`)

**Purpose**: Data Transfer Objects for HTTP requests/responses.

**Request Types**:
- `RegisterRequest` - `{username, password}`
- `LoginRequest` - `{username, password}`
- `UpdateAIPromptRequest` - `{ai_prompt}`
- `UpdateDiscordWebhookRequest` - `{discord_webhook}`
- `UpdateNotifyThresholdRequest` - `{notify_threshold}`

**Response Types**:
- `UserResponse` - User data (without password)
- `LoginResponse` - `{token}`
- `ErrorResponse` - `{code, message}`

**Usage**: Used by HTTP handlers to parse requests and format responses.

---

### Module (`module.go`)

**Purpose**: Registers HTTP routes for the user feature.

**Routes**:
- `POST /api/register` - Public
- `POST /api/login` - Public
- `GET /api/user` - Protected (requires JWT)
- `PUT /api/user/ai-prompt` - Protected
- `PUT /api/user/discord-webhook` - Protected
- `PUT /api/user/notify-threshold` - Protected

**Usage**: Called by router setup in `internal/server/router.go`.

---

### Errors (`usererr/errors.go`)

**Purpose**: User-specific error definitions.

**Errors**:
- `ErrUserAlreadyExists` - Username already taken
- `ErrInvalidCredentials` - Wrong username/password
- `ErrUserNotFound` - User doesn't exist
- `ErrUnauthorized` - Missing/invalid JWT token

**Usage**: Used by service and handler layers for error handling.

---

## Database Schema

**Table**: `users`
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    ai_prompt TEXT,
    discord_webhook TEXT,
    notify_threshold INTEGER DEFAULT 70,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Table**: `user_job_matches`
```sql
CREATE TABLE user_job_matches (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    job_id UUID NOT NULL REFERENCES jobs(id),
    score INTEGER NOT NULL,
    analysis JSONB NOT NULL,
    notified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, job_id)
);
```

**Table**: `preferences`
```sql
CREATE TABLE preferences (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, key)
);
```

---

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `JWT_SECRET` | Yes | Secret key for JWT signing |
| `JWT_EXPIRY_HOURS` | No | Token expiration in hours (default: 24) |

---

## Usage in Pipeline

The user feature is used by other pipeline stages:

1. **User Fanout** (`user_fanout` feature):
   - Calls `UserRepository.GetUsersWithPrompts()` to fetch all users with AI prompts

2. **User Analysis** (`user_analysis` feature):
   - Calls `UserRepository.GetUserByID()` to fetch user
   - Uses `User.AIPrompt` for AI matching
   - Uses `User.NotifyThreshold` to determine if match should be notified

3. **Notification** (`notification` feature):
   - Calls `UserRepository.GetUserByID()` to fetch user
   - Uses `UserJobMatchRepository` to fetch and update matches

---

## Security Considerations

1. **Password Hashing**: Uses bcrypt with default cost
2. **JWT Tokens**: Signed with secret, includes expiration
3. **Protected Routes**: Require valid JWT token
4. **Password Storage**: Never returned in responses

---

## Testing

**Local Testing**:
1. Register user: `POST /api/register`
2. Login: `POST /api/login` (get token)
3. Get user: `GET /api/user` (with token)
4. Update AI prompt: `PUT /api/user/ai-prompt` (with token)

**Production Testing**:
- Test via `api` Lambda endpoints
- Verify JWT token generation and validation
- Test protected routes with/without token

---

## Design Principles

1. **Single Responsibility**: Only handles user management
2. **Security First**: Password hashing, JWT authentication
3. **Reusable**: Repository interfaces used by other features
4. **Observable**: Logs authentication attempts and errors


