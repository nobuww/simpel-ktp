#!/bin/bash
set -e

# ensure .env exists
if [ ! -f .env ]; then
    echo "Error: .env file is missing!"
    echo "Please create one with DATABASE_URL and SESSION_SECRET"
    exit 1
fi

echo "Starting deployment..."

echo "Building and starting services..."
docker-compose up -d --build --remove-orphans web

echo "Waiting for service to report healthy..."
Attempt=0
MaxAttempts=30
while [ $Attempt -lt $MaxAttempts ]; do
    if docker-compose ps web | grep -q "(healthy)"; then
        echo "Service is healthy and running!"
        break
    fi
    echo "Waiting for health check... ($Attempt/$MaxAttempts)"
    sleep 2
    Attempt=$((Attempt+1))
done

if [ $Attempt -eq $MaxAttempts ]; then
    echo "Error: Service failed to pass health check."
    docker-compose logs --tail=50 web
    exit 1
fi

# cleanup
echo "Cleaning up old images..."
docker image prune -f

echo "Deployment successfully completed."
