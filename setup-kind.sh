#!/bin/bash
# Setup script for MathWizz Kind cluster
# Creates cluster and deploys all services

set -e

echo "========================================="
echo "MathWizz Kind Cluster Setup"
echo "========================================="

# Check if kind is installed
if ! command -v kind &> /dev/null; then
    echo "Error: kind is not installed. Please install kind first."
    echo "Visit: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if podman or docker is installed
if command -v podman &> /dev/null; then
    echo "Using podman as container runtime"
    export KIND_EXPERIMENTAL_PROVIDER=podman
    CONTAINER_CMD="podman"
elif command -v docker &> /dev/null; then
    echo "Using docker as container runtime"
    CONTAINER_CMD="docker"
else
    echo "Error: neither podman nor docker is installed. Please install one of them first."
    exit 1
fi

echo "✓ All prerequisites installed"
echo ""

# Create Kind cluster
echo "Creating Kind cluster..."
kind create cluster --config k8s/kind-config.yaml

echo "✓ Kind cluster created"
echo ""

# Build Docker images
echo "Building Docker images..."
bash build-images.sh

echo "✓ Docker images built"
echo ""

# Load images into Kind
echo "Loading images into Kind..."
bash load-images-to-kind.sh

echo "✓ Images loaded into Kind"
echo ""

# Deploy Kubernetes resources
echo "Deploying Kubernetes resources..."

kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secrets.yaml

kubectl apply -f k8s/database-deployment.yaml
kubectl apply -f k8s/database-service.yaml

kubectl apply -f k8s/nats-deployment.yaml
kubectl apply -f k8s/nats-service.yaml

# Wait for database and NATS to be ready
echo "Waiting for database and NATS to be ready..."
kubectl wait --for=condition=ready pod -l app=database -n mathwizz --timeout=120s
kubectl wait --for=condition=ready pod -l app=nats -n mathwizz --timeout=60s

kubectl apply -f k8s/web-server-deployment.yaml
kubectl apply -f k8s/web-server-service.yaml

kubectl apply -f k8s/history-worker-deployment.yaml

kubectl apply -f k8s/frontend-deployment.yaml
kubectl apply -f k8s/frontend-service.yaml

echo "✓ Kubernetes resources deployed"
echo ""

# Wait for all deployments to be ready
echo "Waiting for all deployments to be ready..."
kubectl wait --for=condition=available deployment --all -n mathwizz --timeout=180s

echo ""
echo "========================================="
echo "Setup Complete!"
echo "========================================="
echo ""
echo "Access the application:"
echo "  Frontend: http://localhost:3000"
echo "  API:      http://localhost:8080"
echo ""
echo "View pods:"
echo "  kubectl get pods -n mathwizz"
echo ""
echo "View logs:"
echo "  kubectl logs -f <pod-name> -n mathwizz"
echo ""
echo "To teardown:"
echo "  bash teardown-kind.sh"
echo ""
