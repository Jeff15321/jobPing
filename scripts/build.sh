#!/bin/bash

set -e

echo "Building Go binaries for AWS Lambda..."

cd backend

# Build API (ARM64 for Graviton2 - cheaper and faster)
echo "Building API..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o ../build/bootstrap ./cmd/api

echo "Build complete! Binary is in ./build/"

# Create Lambda deployment package
cd ../build
echo "Creating Lambda deployment package..."

zip -j api.zip bootstrap

echo "Deployment package created!"
echo "- api.zip ($(du -h api.zip | cut -f1))"
