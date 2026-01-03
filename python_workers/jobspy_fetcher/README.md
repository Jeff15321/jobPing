# JobSpy Fetcher Lambda

Fetches jobs from job boards using JobSpy library and sends them to the Go backend for AI analysis.

## Important: Python Version

**Use Python 3.11 or 3.12** - Python 3.13 is not supported due to NumPy compatibility issues.

If you have Python 3.13 installed, you can:
1. Install Python 3.11/3.12 from [python.org](https://www.python.org/downloads/)
2. Create venv with specific version: `py -3.11 -m venv venv` (Windows) or `python3.11 -m venv venv` (Mac/Linux)

## Local Development

### Setup

```bash
# Create virtual environment (use Python 3.11 or 3.12)
py -3.11 -m venv venv
# Activate
source venv/Scripts/activate

# Install dependencies
pip install -r requirements.txt
```

### Test Full Pipeline (Real Jobs)

This tests the complete flow with real job scraping:

```bash
# Make sure Go backend is running (cd backend && air)
python test_full_pipeline.py
```

This will:
1. Scrape real jobs from Indeed using JobSpy
2. Send each job to Go backend via HTTP (`POST /api/jobs/process`)
3. Go backend runs AI analysis (OpenAI or mock)
4. Results saved to PostgreSQL
5. View in frontend at http://localhost:5173

### Test with LocalStack SQS (Optional)

```bash
# Start LocalStack first (from project root: docker-compose up -d)
python test_local.py
```

## Deployment

See main `python_workers/README.md` for deployment instructions.


