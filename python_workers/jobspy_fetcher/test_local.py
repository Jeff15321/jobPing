"""
Local testing script for JobSpy Fetcher.

Run this to test the Lambda handler locally with LocalStack SQS.
"""
import os
import json

# Set environment for local testing
os.environ["SQS_ENDPOINT_URL"] = "http://localhost:4566"
os.environ["SQS_QUEUE_URL"] = "http://localhost:4566/000000000000/jobping-jobs-to-filter"
os.environ["AWS_REGION"] = "us-east-1"
os.environ["AWS_ACCESS_KEY_ID"] = "test"
os.environ["AWS_SECRET_ACCESS_KEY"] = "test"

from handler import handler


def test_fetch_jobs():
    """Test the Lambda handler with a mock API Gateway event."""
    event = {
        "body": json.dumps({
            "search_term": "software engineer",
            "location": "San Francisco, CA",
            "results_wanted": 3
        }),
        "httpMethod": "POST",
        "path": "/jobs/fetch"
    }
    
    print("🚀 Testing JobSpy fetcher...")
    print(f"   Search term: software engineer")
    print(f"   Location: San Francisco, CA")
    print(f"   Results wanted: 3")
    print()
    
    result = handler(event, None)
    
    print("📊 Result:")
    print(f"   Status Code: {result['statusCode']}")
    print(f"   Body: {result['body']}")
    
    return result


if __name__ == "__main__":
    test_fetch_jobs()


