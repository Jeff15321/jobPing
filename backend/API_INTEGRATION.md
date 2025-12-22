# SpeedyApply API Integration Guide

## Current Status

The job fetcher client is implemented in `internal/integrations/jobspy/client.go` but needs to be updated with the actual SpeedyApply API endpoints.

## What We Know

From the SpeedyApply documentation (https://docs.speedyapply.com/discord):
- They have a Discord bot that sends job alerts
- Supports SWE internships, new grad roles, and AI/ML jobs
- Supports both USA and international opportunities

## What We Need

1. **API Endpoint**: The actual REST API endpoint (if public)
2. **Authentication**: API key or token (if required)
3. **Response Format**: JSON structure of job data
4. **Rate Limits**: How often we can poll

## Current Implementation

```go
// File: internal/integrations/jobspy/client.go
func (c *Client) FetchLatestJobs(ctx context.Context, limit int) ([]*job.Job, error) {
    // Currently assumes: GET /jobs?limit=10&type=swe-internship
    url := fmt.Sprintf("%s/jobs?limit=%d&type=swe-internship", c.baseURL, limit)
    // ...
}
```

## Next Steps

### Option 1: Use SpeedyApply Public API (if available)

1. Check if they have a public API at https://api.speedyapply.com
2. Get API documentation
3. Update the client implementation

### Option 2: Scrape Their Website

If no public API exists:
1. Use their job board website
2. Parse HTML or use their internal API calls
3. Be respectful of rate limits

### Option 3: Use Alternative Job APIs

Consider these alternatives:
- **Adzuna API**: https://developer.adzuna.com/
- **The Muse API**: https://www.themuse.com/developers/api/v2
- **GitHub Jobs API**: (deprecated but alternatives exist)
- **Reed API**: https://www.reed.co.uk/developers

## Testing Without Real API

For local development, you can create mock data:

```go
// Add to jobspy/client.go for testing
func (c *Client) FetchLatestJobs(ctx context.Context, limit int) ([]*job.Job, error) {
    // Mock data for testing
    return []*job.Job{
        {
            ID:          "test-1",
            Title:       "Software Engineer Intern",
            Company:     "Google",
            Location:    "Mountain View, CA",
            Description: "Join our team...",
            URL:         "https://careers.google.com/jobs/test-1",
            PostedAt:    time.Now().Add(-24 * time.Hour),
            FetchedAt:   time.Now(),
        },
        // Add more mock jobs...
    }, nil
}
```

## Recommended Approach

1. **Start with mock data** to test the full pipeline
2. **Research SpeedyApply's actual API** or contact them
3. **Implement real integration** once API details are known
4. **Add error handling** for rate limits and failures

## Contact SpeedyApply

If their API isn't documented:
- Check their Discord: https://docs.speedyapply.com/discord
- Look for a support command: `/support`
- Ask about API access for developers

## Alternative: Build Your Own Job Scraper

If SpeedyApply doesn't have a public API, you can:
1. Scrape LinkedIn, Indeed, Glassdoor
2. Use existing job scraping libraries
3. Aggregate from multiple sources

Example libraries:
- **JobSpy** (Python): https://github.com/cullenwatson/JobSpy
- **Indeed Scraper**: Various npm packages
- **LinkedIn API**: Official but limited

## Update Checklist

Once you have the real API details:

- [ ] Update `SPEEDYAPPLY_API_URL` in `.env.example`
- [ ] Update `SpeedyApplyJob` struct with actual fields
- [ ] Update `FetchLatestJobs` with correct endpoint
- [ ] Add authentication if required
- [ ] Add proper error handling
- [ ] Test with real data
- [ ] Update rate limiting logic
