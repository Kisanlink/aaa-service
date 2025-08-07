#!/bin/bash

# AAA Service v2 Build Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="aaa-service"
VERSION=${1:-"v2.0.0"}
DOCKER_IMAGE="kisanlink/aaa-service"
FULL_IMAGE_TAG="${DOCKER_IMAGE}:${VERSION}"
LATEST_TAG="${DOCKER_IMAGE}:latest"

echo -e "${BLUE}Building AAA Service v2...${NC}"
echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo -e "${YELLOW}Image: ${FULL_IMAGE_TAG}${NC}"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

# Clean previous builds
echo -e "${BLUE}Cleaning previous builds...${NC}"
docker system prune -f

# Build the Docker image
echo -e "${BLUE}Building Docker image...${NC}"
docker build -t ${FULL_IMAGE_TAG} -t ${LATEST_TAG} .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Docker image built successfully!${NC}"
    echo -e "${YELLOW}Image tags:${NC}"
    echo -e "  - ${FULL_IMAGE_TAG}"
    echo -e "  - ${LATEST_TAG}"
else
    echo -e "${RED}❌ Docker build failed!${NC}"
    exit 1
fi

# Show image info
echo -e "${BLUE}Image information:${NC}"
docker images | grep ${DOCKER_IMAGE}

# Optional: Push to registry
if [ "$2" = "--push" ]; then
    echo -e "${BLUE}Pushing images to registry...${NC}"
    docker push ${FULL_IMAGE_TAG}
    docker push ${LATEST_TAG}
    echo -e "${GREEN}✅ Images pushed successfully!${NC}"
fi

# Optional: Run tests
if [ "$2" = "--test" ]; then
    echo -e "${BLUE}Running tests...${NC}"
    docker run --rm ${FULL_IMAGE_TAG} go test ./...
    echo -e "${GREEN}✅ Tests completed!${NC}"
fi

echo -e "${GREEN}🎉 AAA Service v2 build completed successfully!${NC}"
echo -e "${YELLOW}To run the service:${NC}"
echo -e "  docker-compose up -d"
echo -e "${YELLOW}To run with custom config:${NC}"
echo -e "  docker run -p 8080:8080 -e DB_HOST=your-db-host ${FULL_IMAGE_TAG}"
