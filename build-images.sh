#!/bin/bash
# Build Docker images for all services
# Builds images locally for use with Kind

set -e

echo "========================================="
echo "Building Docker Images"
echo "========================================="

# Determine container command and build flags
if command -v podman &> /dev/null; then
    CONTAINER_CMD="podman"
    # Use docker format to preserve HEALTHCHECK directives
    BUILD_FLAGS="--format docker"
elif command -v docker &> /dev/null; then
    CONTAINER_CMD="docker"
    BUILD_FLAGS=""
    # Disable BuildKit to ensure images can be saved/loaded properly with Kind
    export DOCKER_BUILDKIT=0
else
    echo "Error: neither podman nor docker is installed"
    exit 1
fi

echo "Using $CONTAINER_CMD to build images"
echo ""

# Build database image
echo "Building database image..."
$CONTAINER_CMD build $BUILD_FLAGS -t mathwizz/database:latest ./database
echo "✓ Database image built"

# Build message-queue image
echo "Building message-queue image..."
$CONTAINER_CMD build $BUILD_FLAGS -t mathwizz/message-queue:latest ./message-queue
echo "✓ Message-queue image built"

# Build web-server image
echo "Building web-server image..."
$CONTAINER_CMD build $BUILD_FLAGS -t mathwizz/web-server:latest ./web-server
echo "✓ Web-server image built"

# Build history-worker image
echo "Building history-worker image..."
$CONTAINER_CMD build $BUILD_FLAGS -t mathwizz/history-worker:latest ./history-worker
echo "✓ History-worker image built"

# Build frontend image
echo "Building frontend image..."
$CONTAINER_CMD build $BUILD_FLAGS -t mathwizz/frontend:latest ./frontend
echo "✓ Frontend image built"

echo ""
echo "All images built successfully!"
