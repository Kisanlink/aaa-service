#!/bin/bash

# Comprehensive Docker cleanup script
echo "Starting comprehensive Docker cleanup..."

# Stop all running containers
echo "Stopping all running containers..."
docker stop $(docker ps -q) 2>/dev/null || echo "No running containers to stop"

# Remove all containers
echo "Removing all containers..."
docker rm $(docker ps -aq) 2>/dev/null || echo "No containers to remove"

# Remove all images
echo "Removing all images..."
docker rmi $(docker images -q) 2>/dev/null || echo "No images to remove"

# Remove all volumes
echo "Removing all volumes..."
docker volume rm $(docker volume ls -q) 2>/dev/null || echo "No volumes to remove"

# Remove all networks (except default ones)
echo "Removing custom networks..."
docker network rm $(docker network ls --filter type=custom -q) 2>/dev/null || echo "No custom networks to remove"

# Prune everything
echo "Pruning Docker system..."
docker system prune -af --volumes

echo "Docker cleanup completed!"
