#!/bin/bash
# Teardown script for MathWizz Kind cluster
# Destroys the Kind cluster and cleans up resources

set -e

echo "========================================="
echo "MathWizz Kind Cluster Teardown"
echo "========================================="

# Check if kind is installed
if ! command -v kind &> /dev/null; then
    echo "Error: kind is not installed."
    exit 1
fi

# Delete the Kind cluster
echo "Deleting Kind cluster..."
kind delete cluster --name mathwizz-cluster 2>&1 | grep -v "using podman due to KIND_EXPERIMENTAL_PROVIDER" | grep -v "enabling experimental podman provider" || true

echo "✓ Kind cluster deleted"
echo ""
echo "Cleanup complete!"
