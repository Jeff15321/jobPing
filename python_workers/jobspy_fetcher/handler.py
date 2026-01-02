"""
JobSpy Fetcher Lambda Handler

Fetches jobs from job boards using JobSpy library and pushes them to SQS.
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


def handler(event: dict, context: Any) -> dict:
    """
    Lambda handler for fetching jobs.
    
    Triggered by API Gateway POST /jobs/fetch
    """
    try:
        # Parse request body
        body = json.loads(event.get("body", "{}")) if event.get("body") else {}
        search_term = body.get("search_term", "software engineer")
        location = body.get("location", "San Francisco, CA")
        results_wanted = body.get("results_wanted", 5)

        logger.info(f"Fetching jobs: term='{search_term}', location='{location}', count={results_wanted}")

        # Fetch jobs using JobSpy
        jobs_df = scrape_jobs(
            site_name=["indeed", "linkedin"],
            search_term=search_term,
            location=location,
            results_wanted=results_wanted,
            hours_old=72,
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

        return {
            "statusCode": 200,
            "headers": {
                "Content-Type": "application/json",
                "Access-Control-Allow-Origin": "*",
            },
            "body": json.dumps({
                "message": f"Fetched {len(jobs_df)} jobs, sent {jobs_sent} to processing queue",
                "jobs_found": len(jobs_df),
                "jobs_queued": jobs_sent,
            }),
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

