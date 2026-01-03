"""
Notification Lambda Handler

Sends Discord notifications via Apprise when users have job matches above their threshold.
"""
import json
import os
import logging
import re
from typing import Any

import apprise

logger = logging.getLogger()
logger.setLevel(logging.INFO)


def convert_discord_webhook_to_apprise(webhook_url: str) -> str:
    """
    Convert Discord webhook URL to Apprise format.
    
    Input: https://discord.com/api/webhooks/123456789/abcdefgh
    Output: discord://123456789/abcdefgh
    """
    pattern = r"https://discord\.com/api/webhooks/(\d+)/([A-Za-z0-9_-]+)"
    match = re.match(pattern, webhook_url)
    if match:
        webhook_id = match.group(1)
        webhook_token = match.group(2)
        return f"discord://{webhook_id}/{webhook_token}"
    
    # If already in apprise format or unknown, return as-is
    return webhook_url


def format_job_notification(message: dict) -> tuple[str, str]:
    """
    Format the job match notification for Discord.
    
    Returns: (title, body)
    """
    job_title = message.get("job_title", "Unknown Position")
    company = message.get("company", "Unknown Company")
    score = message.get("score", 0)
    job_url = message.get("job_url", "")
    
    # Extract analysis details
    analysis = message.get("analysis", {})
    explanation = analysis.get("explanation", "No explanation provided")
    pros = analysis.get("pros", [])
    cons = analysis.get("cons", [])
    
    title = f"Job Match: {job_title} at {company}"
    
    body_parts = [
        f"**Match Score: {score}/100**",
        "",
        explanation,
        "",
    ]
    
    if pros:
        body_parts.append("**Pros:**")
        for pro in pros[:3]:  # Limit to 3
            body_parts.append(f"- {pro}")
        body_parts.append("")
    
    if cons:
        body_parts.append("**Cons:**")
        for con in cons[:3]:  # Limit to 3
            body_parts.append(f"- {con}")
        body_parts.append("")
    
    if job_url:
        body_parts.append(f"[View Job]({job_url})")
    
    body = "\n".join(body_parts)
    
    return title, body


def handler(event: dict, context: Any) -> dict:
    """
    Lambda handler for processing notification SQS messages.
    
    Expected message format:
    {
        "discord_webhook": "https://discord.com/api/webhooks/...",
        "job_title": "Software Engineer",
        "company": "Tech Corp",
        "job_url": "https://...",
        "score": 85,
        "analysis": {
            "explanation": "...",
            "pros": [...],
            "cons": [...]
        }
    }
    """
    notifications_sent = 0
    errors = 0
    
    records = event.get("Records", [])
    logger.info(f"Processing {len(records)} notification requests")
    
    for record in records:
        try:
            message = json.loads(record["body"])
            discord_webhook = message.get("discord_webhook")
            
            if not discord_webhook:
                logger.warning("No discord_webhook in message, skipping")
                continue
            
            # Convert to Apprise format
            apprise_url = convert_discord_webhook_to_apprise(discord_webhook)
            
            # Create Apprise instance
            apobj = apprise.Apprise()
            if not apobj.add(apprise_url):
                logger.error(f"Failed to add notification target: {apprise_url[:50]}...")
                errors += 1
                continue
            
            # Format notification
            title, body = format_job_notification(message)
            
            # Send notification
            result = apobj.notify(
                title=title,
                body=body,
                notify_type=apprise.NotifyType.INFO
            )
            
            if result:
                logger.info(f"Notification sent for job: {message.get('job_title')}")
                notifications_sent += 1
            else:
                logger.error(f"Failed to send notification for job: {message.get('job_title')}")
                errors += 1
                
        except Exception as e:
            logger.error(f"Error processing notification: {e}")
            errors += 1
    
    return {
        "statusCode": 200,
        "body": json.dumps({
            "notifications_sent": notifications_sent,
            "errors": errors
        })
    }


# For local testing
if __name__ == "__main__":
    # Test message
    test_event = {
        "Records": [
            {
                "body": json.dumps({
                    "discord_webhook": os.environ.get("TEST_DISCORD_WEBHOOK", ""),
                    "job_title": "Senior Software Engineer",
                    "company": "Test Company",
                    "job_url": "https://example.com/job/123",
                    "score": 85,
                    "analysis": {
                        "explanation": "This job matches your preferences for remote work and competitive salary.",
                        "pros": ["Remote work", "Good salary range", "Interesting tech stack"],
                        "cons": ["Large company", "Might be slow-moving"]
                    }
                })
            }
        ]
    }
    
    if os.environ.get("TEST_DISCORD_WEBHOOK"):
        result = handler(test_event, None)
        print(f"Result: {result}")
    else:
        print("Set TEST_DISCORD_WEBHOOK environment variable to test")

