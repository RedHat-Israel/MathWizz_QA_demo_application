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
        # With Podman, we need to retag to remove localhost/ prefix,
        # then save to tar and load as archive

        # Clean up any leftover tar file from previous failed runs
        rm -f /tmp/${service}.tar

        # Create tag without localhost/ prefix (overwrites if exists)
        podman tag ${source_image} ${target_image} || {
            echo "Warning: Failed to tag ${source_image} as ${target_image}"
            echo "This might happen if the image was deleted. Rebuilding..."
            return 1
        }

        podman save -o /tmp/${service}.tar ${target_image}
        kind load image-archive /tmp/${service}.tar --name $CLUSTER_NAME 2>&1 | grep -v "using podman due to KIND_EXPERIMENTAL_PROVIDER" | grep -v "enabling experimental podman provider" || true
        rm -f /tmp/${service}.tar

        # Note: We don't clean up the ${target_image} tag - it's harmless and
        # prevents issues with cached builds on subsequent runs
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
