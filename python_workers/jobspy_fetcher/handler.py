"""
JobSpy Fetcher Lambda Handler

Fetches jobs from job boards using JobSpy library and pushes them to SQS.
Supports both API Gateway requests and EventBridge scheduled triggers.
"""
import json
import os
import logging
from typing import Any

import boto3
from jobspy import scrape_jobs

logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Initialize SQS client
sqs = boto3.client("sqs", region_name=os.environ.get("AWS_REGION", "us-east-1"))
SQS_QUEUE_URL = os.environ.get("SQS_QUEUE_URL", "")

# Default settings for scheduled runs
DEFAULT_SEARCH_TERM = "software engineer"
DEFAULT_LOCATION = "United States"
DEFAULT_RESULTS = 10
DEFAULT_HOURS_OLD = 1  # Only fetch jobs from last hour (runs every 30 min)


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

        # Push each job to SQS for AI filtering
        jobs_sent = 0
        for _, job in jobs_df.iterrows():
            job_message = {
                "title": str(job.get("title", "")),
                "company": str(job.get("company", "")),
                "location": str(job.get("location", "")),
                "job_url": str(job.get("job_url", "")),
                "description": str(job.get("description", ""))[:2000],  # Truncate long descriptions
                "job_type": str(job.get("job_type", "")),
                "is_remote": bool(job.get("is_remote", False)),
                "min_amount": float(job.get("min_amount")) if job.get("min_amount") else None,
                "max_amount": float(job.get("max_amount")) if job.get("max_amount") else None,
                "date_posted": str(job.get("date_posted", "")),
            }

            if SQS_QUEUE_URL:
                sqs.send_message(
                    QueueUrl=SQS_QUEUE_URL,
                    MessageBody=json.dumps(job_message),
                )
                jobs_sent += 1
            else:
                logger.warning("SQS_QUEUE_URL not set, skipping SQS")

        result = {
            "message": f"Fetched {len(jobs_df)} jobs, sent {jobs_sent} to processing queue",
            "jobs_found": len(jobs_df),
            "jobs_queued": jobs_sent,
            "trigger": "eventbridge" if event.get("source") == "eventbridge" else "api",
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
