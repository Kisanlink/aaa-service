#!/bin/bash

# Stop SpiceDB and cleanup containers
echo "Stopping SpiceDB and cleaning up..."

# Stop SpiceDB container
if docker ps | grep -q "aaa-spicedb"; then
    echo "Stopping SpiceDB container..."
    docker stop aaa-spicedb
    docker rm aaa-spicedb 2>/dev/null || true
    echo "SpiceDB stopped and removed"
else
    echo "SpiceDB container not running"
fi

# Clean up any dangling containers
echo "Cleaning up dangling containers..."
docker container prune -f

echo "Cleanup completed"
