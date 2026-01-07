#!/bin/bash
# Setup script for MathWizz Kind cluster
# Creates cluster and deploys all services

set -e

# Trap errors and provide helpful message
trap 'echo ""; echo "========================================"; echo "❌ Setup Failed!"; echo "========================================"; echo ""; echo "An error occurred during setup. Check the output above for details."; echo ""; echo "Common issues:"; echo "  - Port 3000 or 8080 already in use"; echo "  - Insufficient disk space for images"; echo "  - Docker/Podman not running"; echo ""; echo "For help, see the Troubleshooting section in README.md"; echo ""; exit 1' ERR

echo "========================================="
echo "MathWizz Kind Cluster Setup"
echo "========================================="
echo ""
echo "Setup Steps:"
echo "  1. Verify prerequisites"
echo "  2. Create/recreate Kind cluster"
echo "  3. Build Docker images"
echo "  4. Load images into Kind"
echo "  5. Generate secrets"
echo "  6. Deploy Kubernetes resources"
echo "  7. Wait for services to be ready"
echo ""
echo "========================================="
echo "Step 1/7: Verifying Prerequisites"
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

echo "========================================="
echo "Step 2/7: Create/Recreate Kind Cluster"
echo "========================================="

# Check if cluster already exists and delete it
CLUSTER_NAME="mathwizz-cluster"
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "⚠️  Existing cluster '${CLUSTER_NAME}' found - deleting for clean setup..."
    kind delete cluster --name $CLUSTER_NAME || true
    echo "✓ Existing cluster deleted"
    echo ""
fi

# Create Kind cluster
echo "Creating Kind cluster..."
kind create cluster --config k8s/kind-config.yaml

echo "✓ Kind cluster created"
echo ""

echo "========================================="
echo "Step 3/7: Building Docker Images"
echo "========================================="
bash build-images.sh

echo ""
echo "========================================="
echo "Step 4/7: Loading Images into Kind"
echo "========================================="
bash load-images-to-kind.sh

echo ""
echo "========================================="
echo "Step 5/7: Generating Secrets"
echo "========================================="

# Generate secrets.yaml if it doesn't exist
if [ ! -f "k8s/secrets.yaml" ]; then
    echo "Generating k8s/secrets.yaml from template..."
    echo "⚠️  WARNING: Using default development password"
    echo "⚠️  FOR DEVELOPMENT ONLY - DO NOT USE IN PRODUCTION"
    echo ""

    # Generate a random password for development
    DEV_PASSWORD="mathwizz_dev_$(date +%s)"
    DB_PASSWORD_B64=$(echo -n "$DEV_PASSWORD" | base64 -w 0 2>/dev/null || echo -n "$DEV_PASSWORD" | base64)

    # Generate a random JWT secret (at least 32 characters)
    # Use date-based fallback if openssl not available
    if command -v openssl &> /dev/null; then
        JWT_SECRET="jwt_secret_$(openssl rand -hex 32)"
    else
        JWT_SECRET="jwt_secret_$(date +%s)_$(echo $RANDOM$RANDOM$RANDOM$RANDOM | md5sum | cut -d' ' -f1)"
    fi
    JWT_SECRET_B64=$(echo -n "$JWT_SECRET" | base64 -w 0 2>/dev/null || echo -n "$JWT_SECRET" | base64)

    # Create secrets.yaml directly (avoids sed delimiter issues with base64)
    cat > k8s/secrets.yaml <<'EOFYAML'
apiVersion: v1
kind: Secret
metadata:
  name: mathwizz-secrets
  namespace: mathwizz
type: Opaque
data:
EOFYAML

    # Append base64 values (prevents issues with special chars in heredoc)
    echo "  DB_PASSWORD: $DB_PASSWORD_B64" >> k8s/secrets.yaml
    echo "  POSTGRES_PASSWORD: $DB_PASSWORD_B64" >> k8s/secrets.yaml
    echo "  JWT_SECRET: $JWT_SECRET_B64" >> k8s/secrets.yaml

    echo "✓ secrets.yaml generated with password: $DEV_PASSWORD"
    echo "  (This file is gitignored and will not be committed)"
else
    echo "✓ Using existing k8s/secrets.yaml"
fi

echo ""
echo "========================================="
echo "Step 6/7: Deploying Kubernetes Resources"
echo "========================================="

kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secrets.yaml

kubectl apply -f k8s/database-deployment.yaml
kubectl apply -f k8s/database-service.yaml

kubectl apply -f k8s/nats-deployment.yaml
kubectl apply -f k8s/nats-service.yaml

echo ""
echo "========================================="
echo "Step 7/7: Waiting for Services to be Ready"
echo "========================================="
echo ""
echo "This may take 2-5 minutes on first run..."
echo ""

# Wait for database and NATS to be ready
echo "[1/3] Waiting for database to be ready..."
if ! kubectl wait --for=condition=ready pod -l app=database -n mathwizz --timeout=300s; then
  echo "ERROR: Database pod failed to become ready"
  echo "Pod status:"
  kubectl get pods -n mathwizz -l app=database
  echo ""
  echo "Pod details:"
  kubectl describe pods -n mathwizz -l app=database
  echo ""
  echo "Pod logs:"
  kubectl logs -n mathwizz -l app=database --tail=50 || echo "No logs available"
  exit 1
fi
echo "✓ Database is ready"
echo ""

echo "[2/3] Waiting for NATS to be ready..."
if ! kubectl wait --for=condition=ready pod -l app=nats -n mathwizz --timeout=180s; then
  echo "ERROR: NATS pod failed to become ready"
  echo "Pod status:"
  kubectl get pods -n mathwizz -l app=nats
  echo ""
  echo "Pod details:"
  kubectl describe pods -n mathwizz -l app=nats
  exit 1
fi
echo "✓ NATS is ready"
echo ""

echo "Deploying application services..."
kubectl apply -f k8s/web-server-deployment.yaml
kubectl apply -f k8s/web-server-service.yaml
kubectl apply -f k8s/history-worker-deployment.yaml
kubectl apply -f k8s/frontend-deployment.yaml
kubectl apply -f k8s/frontend-service.yaml
echo "✓ Application services deployed"
echo ""

# Wait for all deployments to be ready
echo "[3/3] Waiting for application deployments to be ready..."
if ! kubectl wait --for=condition=available deployment --all -n mathwizz --timeout=300s; then
  echo "ERROR: One or more deployments failed to become available"
  echo ""
  echo "Deployment status:"
  kubectl get deployments -n mathwizz
  echo ""
  echo "Pod status:"
  kubectl get pods -n mathwizz
  echo ""
  echo "Failed deployment details:"
  kubectl describe deployments -n mathwizz | grep -A 20 "Replicas:"
  echo ""
  echo "Pod logs for non-ready pods:"
  for pod in $(kubectl get pods -n mathwizz --no-headers | grep -v "1/1.*Running" | awk '{print $1}'); do
    echo "--- Logs for $pod ---"
    kubectl logs -n mathwizz $pod --tail=30 || echo "No logs available"
    echo ""
  done
  exit 1
fi
echo "✓ All deployments are ready!"
echo ""

echo ""
echo "========================================="
echo "✅ Setup Complete!"
echo "========================================="
echo ""
echo "All services are running and ready to use."
echo ""
echo "Pod Status:"
kubectl get pods -n mathwizz
echo ""
echo "Service Status:"
kubectl get svc -n mathwizz
echo ""
echo "========================================="
echo "Access the Application"
echo "========================================="
echo ""
echo "  🌐 Frontend:  http://localhost:3000"
echo "  🔌 API:       http://localhost:8080"
echo "  📊 Health:    http://localhost:8080/health"
echo ""
echo "Quick Start:"
echo "  1. Open http://localhost:3000 in your browser"
echo "  2. Register a new account"
echo "  3. Solve a math problem (e.g., '25+75')"
echo "  4. Check your history"
echo ""
echo "========================================="
echo "Useful Commands"
echo "========================================="
echo ""
echo "View all pods:"
echo "  kubectl get pods -n mathwizz"
echo ""
echo "View logs (follow mode):"
echo "  kubectl logs -f deployment/web-server -n mathwizz"
echo "  kubectl logs -f deployment/history-worker -n mathwizz"
echo "  kubectl logs -f deployment/frontend -n mathwizz"
echo ""
echo "Restart a deployment:"
echo "  kubectl rollout restart deployment/<name> -n mathwizz"
echo ""
echo "Teardown cluster:"
echo "  ./teardown-kind.sh"
echo ""
echo "Capture full output (including errors):"
echo "  ./setup-kind.sh 2>&1 | tee setup-output.txt"
echo ""
echo "========================================="
echo ""
