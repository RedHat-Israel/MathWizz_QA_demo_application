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
    local source_image="${IMAGE_PREFIX}mathwizz/${service}:latest"
    local target_image="mathwizz/${service}:latest"

    echo "Loading ${service} image..."

    if [ "$USE_PODMAN" = true ]; then
        # With Podman, retag to remove localhost/ prefix first
        # This ensures the image exists with the correct name for kind to load
        podman tag ${source_image} ${target_image} 2>/dev/null || {
            echo "Warning: Failed to tag ${source_image} as ${target_image}"
            echo "Image may already be tagged correctly"
        }

        # Use kind load docker-image which works with podman when KIND_EXPERIMENTAL_PROVIDER is set
        # This is more reliable than tar archives for getting the correct image name into Kind
        kind load docker-image ${target_image} --name $CLUSTER_NAME
    else
        # With Docker, use direct docker-image load (requires BuildKit to be disabled)
        # This is more reliable than tar archives when BuildKit is off
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