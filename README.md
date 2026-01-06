# MathWizz - A QA Demo Application

A full-stack math problem-solving application built as a teaching resource for a lecture series about the principles and main working methods in software quality assurance.

Each branch contains the examples of the topics for each lecture in the series talks. **The main branch of this repository does not contain any testing code**.

---


## ⚠️ Important Notice

**This is a teaching tool only, not a production-ready software product.**

MathWizz was created specifically for educational purposes to demonstrate quality assurance concepts and testing methodologies in a realistically-sized and complex codebase. 
The codebase has been vibe-coded to serve as a learning resource and likely contains more bugs than have been discovered or documented by either human or her AI viber.

**Recommended Use**: Use this repository _**solely**_ as a teaching tool for understanding QA principles, testing strategies, and software quality concepts. _**Do not**_ use this code in a production environments or as a foundation for any real-world applications.

---

## Table of Contents

- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
  - [Installation Links](#installation-links)
- [Quick Start](#quick-start)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Setup Kind Cluster](#2-setup-kind-cluster)
  - [3. Access the Application](#3-access-the-application)
  - [4. Test the Application](#4-test-the-application)
- [Configuration](#configuration)
  - [Required Environment Variables](#required-environment-variables)
  - [Optional Environment Variables](#optional-environment-variables)
  - [Math Problem Constraints](#math-problem-constraints)
  - [Cookie-Based Authentication](#cookie-based-authentication)
- [Known Limitations](#known-limitations)
  - [Integer Overflow](#integer-overflow)
  - [Float Truncation](#float-truncation)
  - [Rate Limiting](#rate-limiting)
  - [History Pagination](#history-pagination)
- [Secrets Management](#secrets-management)
  - [Development Environment](#development-environment)
  - [Manual Secret Creation](#manual-secret-creation)
- [Running Tests Locally](#running-tests-locally)
  - [Web-Server Tests](#web-server-tests)
  - [History-Worker Tests](#history-worker-tests)
  - [Frontend Tests](#frontend-tests)
  - [Notes About Running Tests](#notes-about-running-tests)
- [Running Linters](#running-linters)
  - [Go Services](#go-services)
  - [Frontend](#frontend)
- [Project Structure](#project-structure)
- [Development](#development)
  - [Building Individual Services](#building-individual-services)
  - [Viewing Logs](#viewing-logs)
  - [Debugging](#debugging)
- [Teardown](#teardown)
- [CI/CD](#cicd)
  - [Running Workflows Locally with `act`](#running-workflows-locally-with-act)
- [Cross-Platform Support](#cross-platform-support)
- [Troubleshooting](#troubleshooting)
  - [Port Already in Use](#port-already-in-use)
  - [Images Not Loading](#images-not-loading)
  - [Pods Not Starting](#pods-not-starting)
  - [JWT Secret Errors](#jwt-secret-errors)
  - [Database SSL Connection Errors](#database-ssl-connection-errors)
  - [Rate Limit Errors (429 Too Many Requests)](#rate-limit-errors-429-too-many-requests)
  - [CORS Errors](#cors-errors)
- [Contributing](#contributing)
- [License](#license)

## Architecture

The application consists of 5 services:

1. **frontend** (React): User interface with pixel art theme
2. **web-server** (Go): RESTful API server
3. **database** (PostgreSQL): Persistent storage
4. **message-queue** (NATS): Event broker
5. **history-worker** (Go): Async event processor

## Prerequisites

- **Docker** (>= 20.10) - required for Kind cluster and E2E tests
- **Kind** (>= 0.20) - for Kubernetes deployment
- **kubectl** (>= 1.28) - for managing Kubernetes cluster
- **Go** (>= 1.22) - for running backend tests locally
- **Node.js** (>= 18) - for frontend development and E2E tests

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

### 2. Setup MathWizz

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

## Configuration

### Required Environment Variables

The web-server requires the `JWT_SECRET` environment variable to be set:

```bash
# Generate a secure JWT secret
export JWT_SECRET=$(openssl rand -base64 48)

# Or set manually (minimum 32 characters)
export JWT_SECRET="your-secure-secret-key-at-least-32-characters-long"
```

**Important**:
- JWT_SECRET must be at least 32 characters long
- Should be cryptographically random (not a dictionary word or simple phrase)
- The server will fail to start if JWT_SECRET is not set or is too short

**For Kubernetes deployment**, the `setup-kind.sh` script automatically generates a secure JWT_SECRET and includes it in the secrets.yaml file.

### Optional Environment Variables

The following environment variables have defaults and are optional:

| Variable | Default | Description |
|----------|---------|-------------|
| DB_HOST | localhost | PostgreSQL hostname |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | mathwizz | Database username |
| DB_PASSWORD | mathwizz_password | Database password |
| DB_NAME | mathwizz | Database name |
| DB_SSL_MODE | require | SSL encryption mode (disable, require, verify-ca, verify-full) |
| NATS_URL | nats://localhost:4222 | NATS server URL |
| PORT | 8080 | HTTP server port |
| ALLOWED_ORIGINS | http://localhost:3000 | Comma-separated list of allowed CORS origins |

**CORS Configuration**:
- Default origin: `http://localhost:3000` (matches React dev server)
- Multiple origins: separate with commas, no spaces
- Format: `protocol://domain:port` (exact match, case-sensitive)
- Examples:
  - Single domain: `export ALLOWED_ORIGINS="https://mathwizz.com"`
  - Multiple subdomains: `export ALLOWED_ORIGINS="https://mathwizz.com,https://www.mathwizz.com,https://app.mathwizz.com"`
- Security: No wildcards allowed, prevents CSRF attacks

**Database SSL Configuration**:
- Default mode is `require` (encrypted connections)
- For local development without SSL: `export DB_SSL_MODE=disable`
- SSL modes:
  - `disable`: No encryption (INSECURE - local development only)
  - `require`: Encrypted connection (doesn't verify certificate)
  - `verify-ca`: Encrypted + verifies certificate against CA
  - `verify-full`: Encrypted + verifies certificate and hostname (most secure)

### Math Problem Constraints

The `/solve` endpoint enforces security limits to prevent resource exhaustion attacks:

**Input Constraints**:
- Maximum expression length: 100 characters
- Maximum operators: 50 operators (+, -, *, /, ^)
- Maximum nesting depth: 10 levels of parentheses
- Maximum number length: 15 digits per number
- Evaluation timeout: 1 second

**Valid Examples**:
- ✅ `"2+2"` - Simple arithmetic
- ✅ `"25+75"` - Basic addition
- ✅ `"(10+5)*2"` - Parentheses and multiple operations
- ✅ `"100-50+25"` - Multiple operators

**Invalid Examples**:
- ❌ 101-character expression - Too long
- ❌ Expression with 51 operators - Too complex
- ❌ `"((((((((((1+1))))))))))"` (11 levels) - Too deeply nested
- ❌ Numbers with 16+ digits - Number too large

### Cookie-Based Authentication

MathWizz uses **httpOnly cookies** for secure JWT token storage, protecting against XSS (Cross-Site Scripting) attacks.

**How It Works**:
1. When you register or login, the server sets two cookies:
   - `authToken`: Contains your JWT token (HttpOnly=true, inaccessible to JavaScript)
   - `sessionActive`: Session indicator (HttpOnly=false, used by frontend to check login status)

2. Your browser automatically sends these cookies with every API request
3. The server validates the JWT token from the `authToken` cookie
4. You stay logged in for 24 hours (unless you manually logout)

**Cookie Security Attributes**:
- **HttpOnly**: JavaScript cannot read the JWT token (prevents XSS token theft)
- **SameSite=Strict**: Browser only sends cookies to same-origin requests (prevents CSRF attacks)
- **MaxAge=86400**: Cookies expire after 24 hours (matches JWT expiration)
- **Secure=false**: For local development over HTTP

**Why Cookies Instead of localStorage**:
- localStorage is accessible to JavaScript, making tokens vulnerable to XSS attacks
- httpOnly cookies cannot be read by JavaScript, even if an attacker injects malicious scripts
- Browser automatically handles cookie storage and transmission

**For API Clients**:
- Frontend uses `credentials: 'include'` in fetch requests to send cookies automatically
- No need to manually include `Authorization: Bearer <token>` headers
- Backend still accepts `Authorization` header for backward compatibility

**Error Messages**:
When constraints are violated, you'll receive clear error messages:
- `"problem too long: maximum 100 characters allowed"`
- `"expression too complex: maximum 50 operators allowed"`
- `"expression too complex: maximum nesting depth of 10 exceeded"`
- `"number too large: maximum 15 digits allowed"`
- `"evaluation timeout: expression took too long to evaluate"`

## Known Limitations

### Integer Overflow

MathWizz returns integer results limited to the int64 range (-9,223,372,036,854,775,808 to 9,223,372,036,854,775,807).

**⚠️ Warning**: Large calculations may silently overflow and return incorrect results:
- ❌ `999999999999999*999999999999999` returns wrong answer (overflow)
- ❌ `10000000000000*10000000000000` returns wrong answer (overflow)
- ✅ `100000*100000` returns correct answer (10,000,000,000)

**Why This Happens**:
- The 15-digit input limit validates INPUT numbers, not RESULT values
- Two valid 15-digit numbers can multiply to produce results exceeding int64 maximum
- The system does NOT detect overflow - it silently returns an incorrect negative value
- No error is returned when overflow occurs

**Best Practice**: For calculations with very large numbers (>10 trillion), use a dedicated arbitrary-precision calculator instead of MathWizz.

### Float Truncation

Division results are truncated (not rounded) to integers:
- `7/2` returns `3` (not `4`)
- `1/2` returns `0` (not `1`)
- `9/4` returns `2` (internally evaluates to 2.25, then truncates to 2)

**Why This Happens**:
- The solver internally calculates with floating-point precision
- Results are then truncated to integers (decimal part is discarded)
- This matches typical integer division behavior in programming languages

**Workaround**: If you need precise decimal results, use a calculator that supports floating-point output.

### Rate Limiting

The web-server implements rate limiting on authentication endpoints to prevent brute force attacks:

**Endpoints Protected**:
- **POST /register**: 5 requests per minute per IP address
- **POST /login**: 5 requests per minute per IP address

**How It Works**:
- Each IP address gets 5 tokens in a bucket
- Each request consumes 1 token
- Tokens refill at a rate of 1 per 12 seconds (5 per minute)
- When bucket is empty, requests are rejected with HTTP 429

**Rate Limit Response**:
- **Status Code**: 429 Too Many Requests
- **Response Body**: `"rate limit exceeded"`
- **Retry Strategy**: Wait 12 seconds for next token, or 60 seconds for full bucket refill

**Example**:
```bash
# These 5 requests succeed immediately
curl -X POST http://localhost:8080/login  # ✅ Token 1/5
curl -X POST http://localhost:8080/login  # ✅ Token 2/5
curl -X POST http://localhost:8080/login  # ✅ Token 3/5
curl -X POST http://localhost:8080/login  # ✅ Token 4/5
curl -X POST http://localhost:8080/login  # ✅ Token 5/5

# 6th request is rate limited
curl -X POST http://localhost:8080/login  # ❌ 429 Too Many Requests

# Wait 12 seconds, then try again
sleep 12
curl -X POST http://localhost:8080/login  # ✅ Token refilled
```

### History Pagination

The `/history` endpoint supports pagination to efficiently retrieve large history datasets:

**Query Parameters**:
- **limit** (optional): Number of items to return (default: 100, max: 200)
- **offset** (optional): Number of items to skip for pagination (default: 0)

**How It Works**:
- Default behavior returns first 100 items
- Maximum of 200 items per request (prevents memory exhaustion)
- Limit values > 200 are automatically clamped to 200
- Negative offsets are treated as 0
- Results are ordered by `created_at DESC` (newest first)

**Examples**:
```bash
# Get first 100 items (default)
curl -H "Cookie: authToken=YOUR_TOKEN" http://localhost:8080/history

# Get first 50 items
curl -H "Cookie: authToken=YOUR_TOKEN" "http://localhost:8080/history?limit=50"

# Get second page (items 51-100)
curl -H "Cookie: authToken=YOUR_TOKEN" "http://localhost:8080/history?limit=50&offset=50"

# Request 500 items, get 200 (clamped to max)
curl -H "Cookie: authToken=YOUR_TOKEN" "http://localhost:8080/history?limit=500"
```

**Error Handling**:
- Invalid parameters (non-integer) return **400 Bad Request**
- Example: `?limit=abc` → `"invalid limit parameter"`

**Use Cases**:
- **"Load More" button**: Increment offset by limit on each click
- **Infinite scroll**: Fetch next page when user scrolls to bottom
- **Performance**: Limit results when user has thousands of solved problems

## Secrets Management

### Development Environment

The `setup-kind.sh` script automatically generates `k8s/secrets.yaml` from the template with a random development password. This file is **gitignored** and will not be committed to version control.

**⚠️ WARNING**: The auto-generated secrets are for development only and should NEVER be removed from `.gitignore`

### Manual Secret Creation

If you need to manually create secrets:

```bash
# Copy the template
cp k8s/secrets.yaml.template k8s/secrets.yaml

# Generate base64-encoded passwords
echo -n 'your-secure-password' | base64

# Edit k8s/secrets.yaml and replace PLACEHOLDER values with your base64-encoded secrets
```

**Important**: Never commit `k8s/secrets.yaml` to git (it's gitignored). This is a teaching tool - do not use these patterns for any real-world applications.

## Running Tests Locally

All the tests written for this codebase exist in branches `Lecture_02_` and onwards - they've intentionally not been committed to the main branch. 
To view and run them, switch to any of these later branches.

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

# Run E2E tests (fully automated - starts all backend services automatically)
npm run e2e

# Run E2E tests with visible browser
npm run e2e:headed

# Run E2E tests in interactive UI mode
npm run e2e:ui
```

**E2E Test Coverage** (16 tests across 3 files):
- **Login Flow** (`login.e2e.test.js`) - 7 tests: Complete login through UI, error handling, navigation
- **Registration Flow** (`register.e2e.test.js`) - 6 tests: Complete registration through UI, validation, full user journey
- **Solve & History Flow** (`solve.e2e.test.js`) - 3 tests: Problem solving, history retrieval, eventual consistency

### Notes About Running Tests

**E2E Test Setup**:
- E2E tests automatically start all backend services (database, NATS, web-server, history-worker) using Docker Compose
- No manual setup required!
- First run takes 2-3 minutes (builds images), subsequent runs take 30-60 seconds
- See `frontend/e2e/README.md` for details

**Running Tests as CI/CD Workflows**:
- You can also run these tests as part of automated GitHub Actions workflows using the `act` tool
- This simulates the CI/CD pipeline locally in Docker containers
- See the [CI/CD](#cicd) section for instructions on running workflows with `act`

**Speed Up Ginkgo Tests**:
- Add the `-p` flag to `ginkgo` commands to run tests in parallel for faster execution
- **Recommended for unit tests**: `ginkgo -r -p --skip-package=*integration*`
- **⚠️ NOT recommended for integration tests**: `ginkgo -r --focus-file=*integration*` (without `-p`)
- This significantly reduces test runtime by running specs concurrently

**Why avoid `-p` for integration tests?**
- Integration tests often use shared resources (database connections, NATS queues, Docker containers)
- Running them in parallel can cause race conditions, connection pool exhaustion, or port conflicts
- Sequential execution ensures tests don't interfere with each other's state
- Parallel integration test failures are harder to debug due to interleaved output
- The CI/CD pipeline (`.github/workflows/go-ci.yml`) uses `-p` only for unit tests for this reason

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

## CI/CD

**Note**: This repository contains GitHub Actions workflow files (`.github/workflows/*.yml`) as teaching examples to demonstrate CI/CD concepts. These workflows are not actually connected to GitHub Actions since this is a teaching tool, not a production project.

The example workflows demonstrate a typical CI/CD pipeline that would run on every push and pull request:
- Runs all unit and integration tests
- Runs linters for all services
- Builds Docker images
- Reports test coverage

### Running Workflows Locally with `act`

You can test the GitHub Actions workflows locally using [`act`](https://github.com/nektos/act), a tool that runs GitHub Actions workflows in Docker containers on your local machine.

**Installation**:
- Visit https://github.com/nektos/act for installation instructions

**Basic Usage**:

```bash
# List all jobs in the pull_request workflow
act pull_request --list

# Run a specific job
act pull_request -j <job-name>

# Example: Run the test-setup-script job
act pull_request -j test-setup-script

# Run all jobs in the pull_request workflow
act pull_request
```

**Important Notes**:
- Complex workflows (Docker builds, Kind clusters) may be resource-intensive and time-consuming
- At minimum, verify YAML syntax by running `act pull_request --list` - it should show all jobs without errors
- For full integration tests (like `setup-kind.sh`), local testing with `act` is optional but recommended if you have sufficient resources
- `act` uses Docker to simulate the GitHub Actions environment, so ensure Docker is running

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

### JWT Secret Errors

If the web-server fails to start with errors about JWT_SECRET:

**Error**: "JWT_SECRET environment variable is required"
**Solution**: Set the JWT_SECRET environment variable:
```bash
# Generate a secure secret
export JWT_SECRET=$(openssl rand -base64 48)

# Or use a manual secret (min 32 characters)
export JWT_SECRET="your-secure-secret-at-least-32-characters-long"
```

**Error**: "JWT_SECRET must be at least 32 characters long"
**Solution**: Use a longer secret (at least 32 characters):
```bash
# This is too short (will fail)
export JWT_SECRET="short"

# This is valid (32+ characters)
export JWT_SECRET="my-secure-jwt-secret-key-for-development-use-only"
```

**For Kubernetes**: The `setup-kind.sh` script automatically generates a secure JWT_SECRET, so you shouldn't encounter these errors when using the setup script.

### Database SSL Connection Errors

If you encounter SSL-related database connection errors:

**Error**: "SSL not supported" or "failed to ping database: SSL not supported"

**Cause**: Your local PostgreSQL instance doesn't have SSL enabled, but DB_SSL_MODE defaults to "require"

**Solution**: Set DB_SSL_MODE to "disable" for local development:
```bash
# For local development without SSL
export DB_SSL_MODE=disable
```

**Error**: "certificate verify failed" or SSL certificate verification errors

**Cause**: Using verify-ca or verify-full mode without proper certificate configuration

**Solution**: Either use a lower SSL mode or configure certificates:
```bash
# Option 1: Use require mode (doesn't verify certificate)
export DB_SSL_MODE=require

# Option 2: Configure certificates for verify-ca/verify-full
# Mount CA certificate and configure PostgreSQL SSL settings
```

### Rate Limit Errors (429 Too Many Requests)

If you encounter "rate limit exceeded" errors on /login or /register:

**Error**: HTTP 429 with body "rate limit exceeded"

**Cause**: You've exceeded 5 requests per minute from your IP address

**Solutions**:

1. **For Development/Testing**: Wait 12 seconds between requests, or 60 seconds for full reset
   ```bash
   # Wait for token refill
   sleep 12
   curl -X POST http://localhost:8080/login -d '{"email":"test@example.com","password":"pass"}'
   ```

2. **For Automated Tests**: Add delays between authentication requests
   ```javascript
   // Example: Add 12-second delay between login attempts
   await new Promise(resolve => setTimeout(resolve, 12000));
   ```

3. **For Load Testing**: Use multiple IP addresses or test from distributed locations

**Rate Limit Details**:
- Limit: 5 requests per minute per IP
- Refill rate: 1 token per 12 seconds
- Only affects /register and /login endpoints
- Other endpoints (/solve, /history) are not rate limited

**Note**: Rate limits are per-IP and per-server-instance. If you're behind a shared NAT, multiple users may share the same limit.

### CORS Errors

If you encounter CORS errors in the browser console:

**Error**: "Access to fetch at 'http://localhost:8080/...' from origin 'http://localhost:3001' has been blocked by CORS policy"

**Cause**: Your frontend is running on a different origin than what's configured in ALLOWED_ORIGINS

**Solutions**:

1. **Frontend on Different Port**: Update ALLOWED_ORIGINS to match your frontend
   ```bash
   # Frontend on port 3001 instead of default 3000
   export ALLOWED_ORIGINS="http://localhost:3001"

   # Or allow both
   export ALLOWED_ORIGINS="http://localhost:3000,http://localhost:3001"
   ```
   
2. **Multiple Subdomains**: Include all subdomains explicitly
   ```bash
   export ALLOWED_ORIGINS="https://mathwizz.com,https://www.mathwizz.com,https://app.mathwizz.com"
   ```

**Common Mistakes**:
- ❌ Using wildcard: `ALLOWED_ORIGINS="*"` (not supported, security risk)
- ❌ Spaces after commas: `"https://a.com, https://b.com"` (use no spaces)
- ❌ Missing protocol: `"mathwizz.com"` (must include `http://` or `https://`)
- ❌ Wrong case: `"https://MathWizz.com"` vs `"https://mathwizz.com"` (case-sensitive)

**Security Headers**:
The web-server also adds these security headers to all responses:
- `X-Frame-Options: DENY` - Prevents clickjacking (page cannot be embedded in iframe)
- `X-Content-Type-Options: nosniff` - Prevents MIME type sniffing attacks
- `X-XSS-Protection: 1; mode=block` - Enables browser XSS filter

## Contributing

This is a teaching resource and does not accept pull requests at this time. For improvements or bug fixes, please open an issue via the GitHub web console. Please feel free to fork the project for your own use.

## License

This project is for educational purposes.
