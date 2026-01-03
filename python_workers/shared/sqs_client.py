"""
Shared SQS utilities for Python workers.
"""
import json
import os
from typing import Any

import boto3


def get_sqs_client():
    """Get SQS client, supporting LocalStack for local development."""
    endpoint_url = os.environ.get("SQS_ENDPOINT_URL")
    region = os.environ.get("AWS_REGION", "us-east-1")
    
    if endpoint_url:
        # LocalStack or local testing
        return boto3.client(
            "sqs",
            endpoint_url=endpoint_url,
            region_name=region,
            aws_access_key_id="test",
            aws_secret_access_key="test",
        )
    
    return boto3.client("sqs", region_name=region)


def send_message(queue_url: str, message: dict) -> dict:
    """Send a message to SQS queue."""
    client = get_sqs_client()
    return client.send_message(
        QueueUrl=queue_url,
        MessageBody=json.dumps(message),
    )


