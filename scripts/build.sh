#!/bin/bash

set -e

echo "Building Go binaries for AWS Lambda..."

# Verify we're using the correct Go version
GO_VERSION=$(go version)
echo "Using: $GO_VERSION"

cd backend

# Build all Lambda functions
LAMBDAS=("api" "jobs_api" "workers/job_analysis" "workers/user_fanout" "workers/user_analysis" "workers/notifier")

for lambda in "${LAMBDAS[@]}"; do
  name=$(basename $lambda)
  echo "Building $name..."
  GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o ../build/${name}_bootstrap ./cmd/$lambda
done

echo "Build complete! Binaries are in ./build/"

# Create Lambda deployment packages
cd ../build
echo "Creating Lambda deployment packages..."

# Map lambda paths to zip file names
declare -A ZIP_NAMES
ZIP_NAMES["api"]="api.zip"
ZIP_NAMES["jobs_api"]="jobs_api.zip"
ZIP_NAMES["job_analysis"]="job_analysis_worker.zip"
ZIP_NAMES["user_fanout"]="user_fanout_worker.zip"
ZIP_NAMES["user_analysis"]="user_analysis_worker.zip"
ZIP_NAMES["notifier"]="notifier_worker.zip"

# Use PowerShell on Windows, zip on Linux/Mac
for lambda in "${LAMBDAS[@]}"; do
  name=$(basename $lambda)
  zip_name=${ZIP_NAMES[$name]}
  echo "Packaging ${name} as ${zip_name}..."
  mv ${name}_bootstrap bootstrap
  if command -v powershell.exe &> /dev/null; then
    powershell.exe -Command "Compress-Archive -Path bootstrap -DestinationPath ${zip_name} -Force"
  else
    zip -j ${zip_name} bootstrap
  fi
  rm bootstrap
done

echo "Deployment packages created!"
ls -la *.zip 2>/dev/null || echo "No zip files found"
