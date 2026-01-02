"""
Full Pipeline Test - Fetches real jobs and sends to Go backend for AI analysis.

This tests the complete flow locally:
1. JobSpy scrapes real jobs from job boards
2. Jobs are sent to Go backend via HTTP
3. Go backend runs AI analysis
4. Results saved to PostgreSQL

Usage:
    python test_full_pipeline.py
"""
import json
import requests
from jobspy import scrape_jobs

# Go backend URL (local)
GO_BACKEND_URL = "http://localhost:8080"


def fetch_real_jobs(search_term: str = "software engineer", location: str = "San Francisco, CA", count: int = 3):
    """Fetch real jobs using JobSpy."""
    print(f"🔍 Fetching jobs: '{search_term}' in '{location}'...")
    
    try:
        jobs_df = scrape_jobs(
            site_name=["indeed"],  # Start with just Indeed (most reliable)
            search_term=search_term,
            location=location,
            results_wanted=count,
            hours_old=72,
            country_indeed="USA",
        )
        print(f"✅ Found {len(jobs_df)} jobs")
        return jobs_df
    except Exception as e:
        print(f"❌ JobSpy error: {e}")
        return None


def send_job_to_backend(job: dict):
    """Send a single job to Go backend for AI analysis."""
    try:
        response = requests.post(
            f"{GO_BACKEND_URL}/api/jobs/process",
            json=job,
            headers={"Content-Type": "application/json"},
            timeout=30,
        )
        if response.status_code == 200 or response.status_code == 201:
            return response.json()
        else:
            print(f"   ⚠️ Backend returned {response.status_code}: {response.text}")
            return None
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return None


def main():
    print("=" * 60)
    print("🚀 JobPing Full Pipeline Test")
    print("=" * 60)
    print()
    
    # Check if Go backend is running
    try:
        health = requests.get(f"{GO_BACKEND_URL}/health", timeout=5)
        if health.status_code != 200:
            print("❌ Go backend not responding. Start it with: cd backend && air")
            return
        print("✅ Go backend is running")
    except:
        print("❌ Go backend not running. Start it with: cd backend && air")
        return
    
    print()
    
    # Fetch real jobs
    jobs_df = fetch_real_jobs(count=3)
    if jobs_df is None or len(jobs_df) == 0:
        print("❌ No jobs fetched. Check your internet connection.")
        return
    
    print()
    print("📤 Sending jobs to Go backend for AI analysis...")
    print()
    
    processed = 0
    for idx, job in jobs_df.iterrows():
        job_data = {
            "title": str(job.get("title", "")),
            "company": str(job.get("company", "")),
            "location": str(job.get("location", "")),
            "job_url": str(job.get("job_url", "")),
            "description": str(job.get("description", ""))[:2000],
            "job_type": str(job.get("job_type", "")),
            "is_remote": bool(job.get("is_remote", False)),
        }
        
        print(f"   [{idx+1}] {job_data['title']} at {job_data['company']}")
        result = send_job_to_backend(job_data)
        if result:
            processed += 1
            if "ai_score" in result:
                print(f"       → AI Score: {result.get('ai_score')}%")
    
    print()
    print("=" * 60)
    print(f"✅ Pipeline test complete! Processed {processed}/{len(jobs_df)} jobs")
    print()
    print("📊 View results:")
    print(f"   - Frontend: http://localhost:5173")
    print(f"   - API: curl {GO_BACKEND_URL}/api/jobs")
    print("=" * 60)


if __name__ == "__main__":
    main()

