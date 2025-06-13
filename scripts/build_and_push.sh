#!/bin/bash

# Exit on error
set -e

# Configuration
AWS_REGION="ap-south-1"
ECR_REPOSITORY_PREFIX="$(aws sts get-caller-identity --query Account --output text).dkr.ecr.${AWS_REGION}.amazonaws.com"

# Login to ECR
aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${ECR_REPOSITORY_PREFIX}

# Build and push AAA Service
echo "Building AAA Service..."
docker build -t aaa-service:latest .
docker tag aaa-service:latest ${ECR_REPOSITORY_PREFIX}/aaa-service:latest
docker push ${ECR_REPOSITORY_PREFIX}/aaa-service:latest

# Build and push SpiceDB
echo "Building SpiceDB..."
cd ../spicedb  # Adjust path as needed
docker build -t spicedb:latest .
docker tag spicedb:latest ${ECR_REPOSITORY_PREFIX}/spicedb:latest
docker push ${ECR_REPOSITORY_PREFIX}/spicedb:latest

# Build and push Aadhaar Validation Service
echo "Building Aadhaar Validation Service..."
cd ../aadhaar-verification  # Adjust path as needed
docker build -t aadhaar-verification:latest .
docker tag aadhaar-verification:latest ${ECR_REPOSITORY_PREFIX}/aadhaar-verification:latest
docker push ${ECR_REPOSITORY_PREFIX}/aadhaar-verification:latest

echo "All images built and pushed successfully!" 