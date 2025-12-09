#!/usr/bin/env bash

# Function to validate build number (positive integer)
validate_build_number() {
    if ! [[ $1 =~ ^[1-9][0-9]*$ ]]; then
        echo "Error: Build number must be a positive integer"
        echo "Example: 1, 2, 3, etc."
        return 1
    fi
    return 0
}

# Check if build number is provided
if [ -z "$1" ]; then
  echo "Error: Missing build number"
  echo "Usage: ./deploy.sh <build_number>"
  echo "Example: ./deploy.sh 1"
  exit 1
fi

if [ -z "$DASHBOARD_HMAC_SECRET" ]; then
  echo "Error: DASHBOARD_HMAC_SECRET env var is required on host before deploying"
  exit 1
fi

if [ ! -f "config.yaml" ]; then
  echo "Error: config.yaml not found in $(pwd)"
  echo "Place your configuration at ./config.yaml on the host before deploying"
  exit 1
fi

BUILD_NUMBER=$1
IMAGE_NAME="kovadocker/dashboard-backend"
CONTAINER_NAME="dashboard-backend"

# Validate input
if ! validate_build_number "$BUILD_NUMBER"; then
    exit 1
fi

echo "Deploying $IMAGE_NAME with build: $BUILD_NUMBER"

# Pull new image
echo "Pulling new image..."
if ! docker pull $IMAGE_NAME:$BUILD_NUMBER; then
    echo "Error: Failed to pull new image"
    exit 1
fi

# Create or replace the container using docker compose
docker compose -f - up -d <<EOF
services:
  $CONTAINER_NAME:
    image: $IMAGE_NAME:$BUILD_NUMBER
    container_name: $CONTAINER_NAME
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - DASHBOARD_HMAC_SECRET=${DASHBOARD_HMAC_SECRET}
    volumes:
      - "$(pwd)/config.yaml:/home/config.yaml:ro"
    networks:
      - mynet

networks:
  mynet:
    driver: bridge
EOF

if [ $? -ne 0 ]; then
  echo "Error: Docker Compose failed"
  exit 1
fi

# Verify the container is running
if ! docker ps | grep -q "$CONTAINER_NAME"; then
    echo "Error: Container failed to start"
    echo "Checking container logs:"
    docker logs "$CONTAINER_NAME"
    exit 1
fi

echo "Deployment complete!" 

