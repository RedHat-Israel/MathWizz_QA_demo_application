# MathWizz - A QA Demo Application

A full-stack math problem-solving application built as a teaching resource for a lecture series about the principles and main working methods in software quality assurance.

Each branch contains the examples of the topics for each lecture in the series talks.

## Architecture

The application consists of 5 services:

1. **frontend** (React): User interface with pixel art theme
2. **web-server** (Go): RESTful API server
3. **database** (PostgreSQL): Persistent storage
4. **message-queue** (NATS): Event broker
5. **history-worker** (Go): Async event processor

## Prerequisites

- **Docker** (>= 20.10)
- **Kind** (>= 0.20)
- **kubectl** (>= 1.28)
- **Go** (>= 1.22) - for running tests locally
- **Node.js** (>= 18) - for frontend development

### Installation Links

- Docker: https://docs.docker.com/get-docker/
- Kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installation
- kubectl: https://kubernetes.io/docs/tasks/tools/
- Go: https://golang.org/doc/install
- Node.js: https://nodejs.org/

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd MathWizz
```

### 2. Setup Kind Cluster

```bash
chmod +x setup-kind.sh
./setup-kind.sh
```

This script will:
- Create a Kind cluster
- Build all Docker images
- Load images into Kind
- Deploy all services to Kubernetes

### 3. Access the Application

- **Frontend**: http://localhost:3000
- **API**: http://localhost:8080

### 4. Test the Application

1. Register a new account
2. Login with your credentials
3. Solve a math problem (e.g., "25+75")
4. Check your history to see the solved problem

## Running Tests Locally - MAKE SURE YOU'RE USING A BRANCH THAT HAS TESTS!

### Web-Server Tests

```bash
cd web-server

# Run all tests
ginkgo -r

# Run unit tests only
ginkgo -r --skip-package=*integration*

# Run integration tests only
ginkgo -r --focus-file=*integration*
```

### History-Worker Tests

```bash
cd history-worker

# Run all tests
ginkgo -r

# Run unit tests only
ginkgo -r --skip-package=*integration*

# Run integration tests only
ginkgo -r --focus-file=*integration*
```

### Frontend Tests

```bash
cd frontend

# Run unit and component tests
npm test

# Run E2E tests (requires running backend)
npm run e2e
```

## Running Linters

### Go Services

```bash
# Web-server
cd web-server
golangci-lint run

# History-worker
cd history-worker
golangci-lint run
```

### Frontend

```bash
cd frontend
npm run lint
```

## Project Structure

```
MathWizz/
├── frontend/           # React application
├── web-server/         # Go API server
├── database/           # PostgreSQL configuration
├── message-queue/      # NATS configuration
├── history-worker/     # Go async worker
├── k8s/                # Kubernetes manifests
└── scripts/            # Deployment scripts
```

## Development

### Building Individual Services

```bash
# Database
docker build -t mathwizz/database:latest ./database

# Message Queue
docker build -t mathwizz/message-queue:latest ./message-queue

# Web-Server
docker build -t mathwizz/web-server:latest ./web-server

# History-Worker
docker build -t mathwizz/history-worker:latest ./history-worker

# Frontend
docker build -t mathwizz/frontend:latest ./frontend
```

### Viewing Logs

```bash
# List all pods
kubectl get pods -n mathwizz

# View logs for a specific pod
kubectl logs -f <pod-name> -n mathwizz

# View web-server logs
kubectl logs -f deployment/web-server -n mathwizz

# View history-worker logs
kubectl logs -f deployment/history-worker -n mathwizz
```

### Debugging

```bash
# Describe a pod
kubectl describe pod <pod-name> -n mathwizz

# Get events
kubectl get events -n mathwizz --sort-by='.lastTimestamp'

# Execute commands in a pod
kubectl exec -it <pod-name> -n mathwizz -- /bin/sh
```

## Teardown

To delete the Kind cluster and all resources:

```bash
chmod +x teardown-kind.sh
./teardown-kind.sh
```

## Testing Strategy

See [TESTING.md](TESTING.md) for detailed information about the testing approach.

## Architecture Details

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed architecture documentation.

## CI/CD

GitHub Actions pipeline runs on every push and pull request:
- Runs all unit and integration tests
- Runs linters for all services
- Builds Docker images
- Reports test coverage

## Cross-Platform Support

This project is designed to work on:
- **Linux** (Fedora CSB OS and other distributions)
- **macOS** (Intel and Apple Silicon)

## Troubleshooting

### Port Already in Use

If ports 3000 or 8080 are already in use, modify the `kind-config.yaml` port mappings.

### Images Not Loading

Ensure Docker is running and you have built the images:

```bash
./build-images.sh
./load-images-to-kind.sh
```

### Pods Not Starting

Check pod status and logs:

```bash
kubectl get pods -n mathwizz
kubectl describe pod <pod-name> -n mathwizz
kubectl logs <pod-name> -n mathwizz
```

## Contributing

This is a teaching resource. For improvements or bug fixes, please create a pull request.

## License

This project is for educational purposes.
