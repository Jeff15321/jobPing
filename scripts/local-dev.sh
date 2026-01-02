#!/bin/bash
# Local development startup script

set -e

echo "🚀 Starting JobPing local development environment..."

# Start Docker services
echo "📦 Starting Docker services (Postgres + LocalStack)..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be healthy..."
sleep 5

# Check if services are ready
docker-compose ps

echo ""
echo "✅ Services started!"
echo ""
echo "📝 Quick reference:"
echo "   - Postgres: localhost:5433"
echo "   - LocalStack SQS: http://localhost:4566"
echo "   - SQS Queue URL: http://localhost:4566/000000000000/jobping-jobs-to-filter"
echo ""
echo "🔧 To start the Go backend:"
echo "   cd backend && air"
echo ""
echo "🔧 To start the frontend:"
echo "   cd frontend && npm run dev"
echo ""
echo "🐍 To test Python worker locally:"
echo "   cd python_workers/jobspy_fetcher"
echo "   pip install -r requirements.txt"
echo "   python test_local.py"
echo ""
