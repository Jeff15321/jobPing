#!/bin/bash

echo "Testing API..."
echo ""

echo "1. Health check:"
curl -s http://localhost:8080/health | jq .
echo ""

echo "2. Get first job:"
curl -s http://localhost:8080/api/job | jq .
echo ""

echo "Done!"
