#!/bin/bash

set -e

echo "Building Go binaries for AWS Lambda..."

cd backend

# Build API
echo "Building API..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../build/api ./cmd/api

# Build Scanner
echo "Building Scanner..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../build/scanner ./cmd/scanner

# Build Matcher
echo "Building Matcher..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../build/matcher ./cmd/matcher

echo "Build complete! Binaries are in ./build/"

# Create Lambda deployment packages
cd ../build
echo "Creating Lambda deployment packages..."

zip -j api.zip api
zip -j scanner.zip scanner
zip -j matcher.zip matcher

echo "Deployment packages created!"
echo "- api.zip"
echo "- scanner.zip"
echo "- matcher.zip"
