#!/bin/bash

set -e

echo "Building Go binaries for AWS Lambda..."


# Verify we're using the correct Go version
GO_VERSION=$(go version)
echo "Using: $GO_VERSION"

cd backend

# Build Lambda function (ARM64 for Graviton2 - cheaper and faster)
echo "Building Lambda function..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o ../build/bootstrap ./cmd/lambda

echo "Build complete! Binary is in ./build/"

# Create Lambda deployment package
cd ../build
echo "Creating Lambda deployment package..."

# Use PowerShell on Windows, zip on Linux/Mac
if command -v powershell.exe &> /dev/null; then
  powershell.exe -Command "Compress-Archive -Path bootstrap -DestinationPath api.zip -Force"
else
  zip -j api.zip bootstrap
fi

echo "Deployment package created!"
if [ -f api.zip ]; then
  echo "- api.zip ($(du -h api.zip | cut -f1))"
else
  echo "Warning: api.zip was not created"
fi
