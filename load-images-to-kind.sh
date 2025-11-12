#!/bin/bash
# Load Docker images into Kind cluster
# Makes locally built images available to Kind

set -e

echo "========================================="
echo "Loading Images into Kind"
echo "========================================="

CLUSTER_NAME="mathwizz-cluster"

# Determine container runtime and image prefix
USE_PODMAN=false
IMAGE_PREFIX=""
if command -v podman &> /dev/null && [ -n "$KIND_EXPERIMENTAL_PROVIDER" ]; then
    USE_PODMAN=true
    IMAGE_PREFIX="localhost/"
fi

# Function to load image
load_image() {
    local service=$1
    local image_name="${IMAGE_PREFIX}mathwizz/${service}:latest"

    echo "Loading ${service} image..."

    if [ "$USE_PODMAN" = true ]; then
        # With Podman, save to tar and load as archive
        podman save -o /tmp/${service}.tar ${image_name}
        kind load image-archive /tmp/${service}.tar --name $CLUSTER_NAME
        rm /tmp/${service}.tar
    else
        # With Docker, use direct load
        kind load docker-image mathwizz/${service}:latest --name $CLUSTER_NAME
    fi

    echo "✓ ${service} image loaded"
}

# Load all images
load_image "database"
load_image "message-queue"
load_image "web-server"
load_image "history-worker"
load_image "frontend"

echo ""
echo "All images loaded into Kind successfully!"
