#!/bin/bash

echo "🚀 Starting AI Job Scanner - Local Development"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Start Docker services
echo "📦 Starting PostgreSQL and LocalStack..."
docker-compose up -d postgres localstack

echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+"
    exit 1
fi

# Check if Node is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+"
    exit 1
fi

echo ""
echo "✅ Services are ready!"
echo ""
echo "Next steps:"
echo "1. Start the API server:"
echo "   cd backend && go run cmd/api/main.go"
echo ""
echo "2. In another terminal, start the frontend:"
echo "   cd frontend && npm install && npm run dev"
echo ""
echo "3. (Optional) Run the scanner to fetch jobs:"
echo "   cd backend && go run cmd/scanner/main.go"
echo ""
echo "📱 Frontend: http://localhost:5173"
echo "🔌 API: http://localhost:8080"
echo "💾 Database: localhost:5432"
