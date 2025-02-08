#!/bin/bash

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

BUILD_NUMBER=$1
IMAGE_NAME="kovadocker/dashboard-backend"
CONTAINER_NAME="dashboard-backend"

# Validate input
if ! validate_build_number "$BUILD_NUMBER"; then
    exit 1
fi

# Log in to docker
echo "Docker login..."
if ! echo "${{ secrets.WERIK_DOCKER_HUB_TOKEN }}" | docker login -u ${{ secrets.DOCKER_HUB_USERNAME }} --password-stdin >/dev/null 2>&1; then
    echo "Error: Failed to log in to Docker Hub"
    exit 1
fi

# Check if image exists in registry
echo "Checking if image exists..."
if ! docker manifest inspect $IMAGE_NAME:$BUILD_NUMBER >/dev/null 2>&1; then
    echo "Error: Image $IMAGE_NAME:$BUILD_NUMBER does not exist"
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
      - "8888:8080"
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

