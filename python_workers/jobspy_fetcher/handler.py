"""
JobSpy Fetcher Lambda Handler

Fetches jobs from job boards using JobSpy library, stores them in RDS, and sends job_id to SQS.
Supports both API Gateway requests and EventBridge scheduled triggers.
"""
import json
import os
import logging
from typing import Any
from datetime import datetime
import uuid

import boto3
import psycopg2
from jobspy import scrape_jobs

logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Initialize SQS client
sqs = boto3.client("sqs", region_name=os.environ.get("AWS_REGION", "us-east-1"))
SQS_QUEUE_URL = os.environ.get("JOB_ANALYSIS_QUEUE_URL", "")
DATABASE_URL = os.environ.get("DATABASE_URL", "")

# Default settings for scheduled runs
DEFAULT_SEARCH_TERM = "software engineer"
DEFAULT_LOCATION = "United States"
DEFAULT_RESULTS = 10
DEFAULT_HOURS_OLD = 1  # Only fetch jobs from last hour (runs every 30 min)


def get_db_connection():
    """Get PostgreSQL database connection."""
    if not DATABASE_URL:
        return None
    try:
        return psycopg2.connect(DATABASE_URL)
    except Exception as e:
        logger.error(f"Failed to connect to database: {e}")
        return None


def parse_event(event: dict) -> dict:
    """
    Parse event from either API Gateway or EventBridge.
    
    API Gateway: event has 'body' field with JSON string
    EventBridge: event has direct fields (search_term, location, etc.)
    """
    # Check if this is an EventBridge scheduled event
    if event.get("source") == "eventbridge" or "detail-type" in event:
        logger.info("Processing EventBridge scheduled event")
        return {
            "search_term": event.get("search_term", DEFAULT_SEARCH_TERM),
            "location": event.get("location", DEFAULT_LOCATION),
            "results_wanted": event.get("results_wanted", DEFAULT_RESULTS),
            "hours_old": event.get("hours_old", DEFAULT_HOURS_OLD),
        }
    
    # Check if this is an API Gateway event
    if event.get("body"):
        logger.info("Processing API Gateway event")
        body = json.loads(event.get("body", "{}"))
        return {
            "search_term": body.get("search_term", DEFAULT_SEARCH_TERM),
            "location": body.get("location", DEFAULT_LOCATION),
            "results_wanted": body.get("results_wanted", 5),
            "hours_old": body.get("hours_old", 72),  # API calls get more history
        }
    
    # Direct invocation or test
    logger.info("Processing direct invocation")
    return {
        "search_term": event.get("search_term", DEFAULT_SEARCH_TERM),
        "location": event.get("location", DEFAULT_LOCATION),
        "results_wanted": event.get("results_wanted", DEFAULT_RESULTS),
        "hours_old": event.get("hours_old", DEFAULT_HOURS_OLD),
    }


def handler(event: dict, context: Any) -> dict:
    """
    Lambda handler for fetching jobs.
    
    Triggered by:
    - API Gateway POST /jobs/fetch
    - EventBridge scheduled rule (every 30 minutes)
    """
    try:
        # Parse event parameters
        params = parse_event(event)
        search_term = params["search_term"]
        location = params["location"]
        results_wanted = params["results_wanted"]
        hours_old = params["hours_old"]

        logger.info(f"Fetching jobs: term='{search_term}', location='{location}', count={results_wanted}, hours_old={hours_old}")

        # Fetch jobs using JobSpy
        jobs_df = scrape_jobs(
            site_name=["indeed", "linkedin", "glassdoor"],
            search_term=search_term,
            location=location,
            results_wanted=results_wanted,
            hours_old=hours_old,
            country_indeed="USA",
        )

        logger.info(f"Found {len(jobs_df)} jobs")

        # Connect to database
        conn = get_db_connection()
        if not conn:
            logger.error("Database connection failed, cannot store jobs")
            return {
                "statusCode": 500,
                "headers": {
                    "Content-Type": "application/json",
                    "Access-Control-Allow-Origin": "*",
                },
                "body": json.dumps({"error": "Database connection failed"}),
            }

        jobs_created = 0
        jobs_queued = 0
        all_jobs = []  # Store all fetched jobs for response
        
        try:
            cur = conn.cursor()
            now = datetime.utcnow()
            
            for _, job_row in jobs_df.iterrows():
                job_id = uuid.uuid4()
                job_url = str(job_row.get("job_url", ""))
                
                # Collect job data for response (before checking if exists)
                job_data = {
                    "id": str(job_id),
                    "title": str(job_row.get("title", "")),
                    "company": str(job_row.get("company", "")),
                    "location": str(job_row.get("location", "")),
                    "job_url": job_url,
                    "job_type": str(job_row.get("job_type", "")),
                    "is_remote": bool(job_row.get("is_remote", False)),
                    "min_salary": float(job_row.get("min_amount")) if job_row.get("min_amount") else None,
                    "max_salary": float(job_row.get("max_amount")) if job_row.get("max_amount") else None,
                    "date_posted": str(job_row.get("date_posted", "")),
                    "status": "existing",  # Default to existing, will change if created
                }
                
                # Check if job already exists
                cur.execute("SELECT id FROM jobs WHERE job_url = %s", (job_url,))
                existing = cur.fetchone()
                
                if existing:
                    logger.info(f"Job already exists: {job_url}, skipping")
                    all_jobs.append(job_data)
                    continue
                
                # Insert job into database
                cur.execute("""
                    INSERT INTO jobs (
                        id, title, company, location, job_url, description, job_type,
                        is_remote, min_salary, max_salary, date_posted, status,
                        created_at, updated_at
                    ) VALUES (
                        %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s
                    )
                """, (
                    str(job_id),
                    str(job_row.get("title", "")),
                    str(job_row.get("company", "")),
                    str(job_row.get("location", "")),
                    job_url,
                    str(job_row.get("description", ""))[:5000],  # Truncate if too long
                    str(job_row.get("job_type", "")),
                    bool(job_row.get("is_remote", False)),
                    float(job_row.get("min_amount")) if job_row.get("min_amount") else None,
                    float(job_row.get("max_amount")) if job_row.get("max_amount") else None,
                    str(job_row.get("date_posted", "")),
                    "pending",
                    now,
                    now,
                ))
                
                jobs_created += 1
                job_data["status"] = "created"  # Mark as newly created
                all_jobs.append(job_data)
                
                # Send job_id to SQS
                if SQS_QUEUE_URL:
                    message = {"job_id": str(job_id)}
                    sqs.send_message(
                        QueueUrl=SQS_QUEUE_URL,
                        MessageBody=json.dumps(message),
                    )
                    jobs_queued += 1
                else:
                    logger.warning("JOB_ANALYSIS_QUEUE_URL not set, skipping SQS")
            
            conn.commit()
            cur.close()
            
        except Exception as e:
            logger.error(f"Error processing jobs: {e}")
            conn.rollback()
            raise
        finally:
            conn.close()

        result = {
            "message": f"Fetched {len(jobs_df)} jobs, created {jobs_created} in database, queued {jobs_queued} for analysis",
            "jobs_found": len(jobs_df),
            "jobs_created": jobs_created,
            "jobs_queued": jobs_queued,
            "trigger": "eventbridge" if event.get("source") == "eventbridge" else "api",
            "jobs": all_jobs,  # Include all fetched jobs (both new and existing)
        }

        # Return API Gateway formatted response
        return {
            "statusCode": 200,
            "headers": {
                "Content-Type": "application/json",
                "Access-Control-Allow-Origin": "*",
            },
            "body": json.dumps(result),
        }

    except Exception as e:
        logger.error(f"Error fetching jobs: {e}")
        return {
            "statusCode": 500,
            "headers": {
                "Content-Type": "application/json",
                "Access-Control-Allow-Origin": "*",
            },
            "body": json.dumps({"error": str(e)}),
        }
