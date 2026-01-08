# Feature Issues and Recommendations

This document lists issues found during code review and recommendations for cleanup.

## Critical Issues (Should Fix)

### 1. Unused Handler Files in `job` Feature

**Files to Delete**:
- `backend/internal/features/job/handler/sqs.go`
- `backend/internal/features/job/handler/job_analysis.go`

**Reason**: These are old handlers that have been replaced by the new pipeline features:
- `job/handler/sqs.go` → Replaced by `job_analysis/handler/sqs.go`
- `job/handler/job_analysis.go` → Replaced by `job_analysis/handler/sqs.go`

**Impact**: Dead code that can cause confusion.

**Action**: Delete these files.

---

### 2. Legacy SQS Client Uses Old Queue

**File**: `backend/internal/features/job/service/sqs_client.go`

**Issue**: Uses environment variable `NOTIFY_SQS_QUEUE_URL` which is the old queue name. The new pipeline uses `NOTIFICATION_QUEUE_URL`.

**Current Code**:
```go
queueURL := os.Getenv("NOTIFY_SQS_QUEUE_URL")
```

**Should Be**:
```go
queueURL := os.Getenv("NOTIFICATION_QUEUE_URL")
```

**Impact**: If `JobService` is used in production (it shouldn't be), it would use the wrong queue.

**Action**: Update environment variable name or remove if `JobService` is local-only.

---

## Medium Priority Issues (Consider Fixing)

### 3. JobService Does Too Much

**File**: `backend/internal/features/job/service/job_service.go`

**Issue**: The `ProcessJob` method performs all pipeline stages in one function:
1. Company research
2. AI analysis
3. User matching
4. Notification queuing

**Problem**: This violates single responsibility and duplicates logic now in separate features.

**Current Usage**: 
- Used by `job/handler/mock.go` for local testing
- Used by `job/handler/http.go` for local testing

**Recommendation**: 
- Keep for local dev but mark as deprecated
- Consider simplifying to just CRUD operations
- Or split into smaller methods that call the new feature services

**Action**: Add deprecation comment, consider refactoring.

---

### 4. AI Client Duplication

**File**: `backend/internal/features/job/service/ai_client.go`

**Issue**: This AI client duplicates functionality now in separate features:
- `job/service/ai_client.go` - Full interface (AnalyzeJob, ResearchCompany, MatchJobToUser)
- `job_analysis/service/ai_client.go` - Only ResearchCompany
- `user_analysis/service/ai_client.go` - Only MatchJobToUser

**Problem**: Code duplication and maintenance burden.

**Recommendation**: 
- Remove once `JobService` is simplified
- Or extract to shared package if needed

**Action**: Mark as deprecated, remove when `JobService` is refactored.

---

## Low Priority Issues (Nice to Have)

### 5. Notification Idempotency

**File**: `backend/internal/features/notification/service/notification_service.go`

**Issue**: The `SendNotification` method doesn't check if a notification already exists for a match. It will create duplicate notifications if called multiple times.

**Recommendation**: Add check to prevent duplicate notifications:
```go
existing, err := s.notifRepo.GetByMatchID(ctx, matchID)
if existing != nil {
    log.Printf("Notification already exists for match %s", matchID)
    return nil
}
```

**Action**: Add duplicate check (optional, low priority).

---

### 6. Error Handling in Fanout

**File**: `backend/internal/features/user_fanout/service/fanout_service.go`

**Issue**: If individual user enqueue fails, it logs and continues. This is fine, but there's no way to retry failed users without re-running the entire fanout.

**Recommendation**: Consider adding retry logic or DLQ for individual user failures (optional).

**Action**: Current behavior is acceptable, but could be improved.

---

### 7. Missing Batch Operations

**File**: `backend/internal/features/user_fanout/service/fanout_service.go`

**Issue**: Enqueues messages one at a time. Could be optimized with batch operations.

**Recommendation**: Use `SendMessageBatch` for better performance (optional optimization).

**Action**: Current behavior works, but could be optimized.

---

## Summary

### Immediate Actions
1. ✅ Delete `job/handler/sqs.go`
2. ✅ Delete `job/handler/job_analysis.go`
3. ⚠️ Update `job/service/sqs_client.go` to use `NOTIFICATION_QUEUE_URL`

### Future Refactoring
1. Simplify `JobService` or mark as deprecated
2. Remove `job/service/ai_client.go` once `JobService` is simplified
3. Add notification idempotency check (optional)

### Optional Optimizations
1. Batch SQS operations in fanout
2. Add retry logic for individual user failures

---

## Files That Should Be Removed

1. `backend/internal/features/job/handler/sqs.go` - Replaced by `job_analysis/handler/sqs.go`
2. `backend/internal/features/job/handler/job_analysis.go` - Replaced by `job_analysis/handler/sqs.go`

## Files That Should Be Updated

1. `backend/internal/features/job/service/sqs_client.go` - Update env var name
2. `backend/internal/features/job/service/job_service.go` - Add deprecation comment

## Files That Could Be Removed (Future)

1. `backend/internal/features/job/service/ai_client.go` - Once `JobService` is simplified


