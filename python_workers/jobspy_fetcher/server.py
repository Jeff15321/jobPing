"""
Flask server for local development.
Wraps the Lambda handler for HTTP access.
"""
import json
import os
from flask import Flask, request, jsonify
from handler import handler

app = Flask(__name__)


@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint."""
    return jsonify({"status": "ok"})


@app.route('/fetch', methods=['POST'])
def fetch_jobs():
    """
    HTTP endpoint for local development.
    Wraps the Lambda handler.
    """
    data = request.json or {}
    
    # Build event structure that handler expects
    event = {
        'body': json.dumps({
            'search_term': data.get('search_term', 'software engineer'),
            'location': data.get('location', 'San Francisco, CA'),
            'results_wanted': data.get('results_wanted', 5),
            'hours_old': data.get('hours_old', 72),
        })
    }
    
    # Call the Lambda handler
    result = handler(event, None)
    
    # Parse the response
    status_code = result.get('statusCode', 500)
    body = json.loads(result.get('body', '{}'))
    
    return jsonify(body), status_code


if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8081))
    # debug=True enables hot reload
    app.run(host='0.0.0.0', port=port, debug=True)
