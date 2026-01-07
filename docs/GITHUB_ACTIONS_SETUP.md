# GitHub Actions & CI/CD Setup Summary

This document provides an overview of the GitHub Actions workflows and CI/CD setup for the Event-Driven Order System.

## Files Added

### 1. GitHub Actions Workflows

#### `.github/workflows/ci.yml`
Main CI workflow that runs on every push and pull request to `main` and `develop` branches.

**Jobs:**
- **Lint**: Runs `golangci-lint` on all three services
- **Test**: Runs unit tests with race detection and coverage reporting
- **Build**: Builds Go binaries for all services
- **Integration**: Spins up MySQL, Redis, and RabbitMQ for integration testing

**Matrix Strategy:** Runs jobs in parallel for:
- order-service
- analytics-service
- notification-worker

#### `.github/workflows/docker.yml`
Docker build and validation workflow.

**Jobs:**
- **docker-build**: Builds Docker images for each service with BuildKit caching
- **docker-compose**: Validates docker-compose.yml and tests the full stack

### 2. Linter Configuration

#### `.golangci.yml`
Comprehensive Go linting configuration with:
- 20+ enabled linters
- Custom rules per linter
- Test file exceptions
- Performance and security checks
- Code complexity limits
- Import formatting

**Key Features:**
- Max line length: 120 characters
- Max function lines: 100
- Max cyclomatic complexity: 15
- Security checks with `gosec`
- Code duplication detection
- Unused code detection

### 3. Health Check Endpoints

Enhanced health check endpoints for all services that check:
- Database connectivity
- Redis cache connectivity
- RabbitMQ message queue connectivity

**Endpoints:**
- Order Service: `http://localhost:8080/health`
- Analytics Service: `http://localhost:8081/health`
- Notification Worker: `http://localhost:8082/health`

**Response Format:**
```json
{
  "status": "healthy",
  "service": "order-service",
  "checks": {
    "database": "healthy",
    "cache": "healthy",
    "mq": "healthy"
  }
}
```

**Health States:**
- `200 OK`: All checks passed, service is healthy
- `503 Service Unavailable`: One or more checks failed, service is degraded

### 4. Branch Protection

#### `docs/BRANCH_PROTECTION.md`
Comprehensive guide for setting up branch protection rules with:
- Web interface instructions
- GitHub CLI commands
- Automated setup script
- Verification steps
- Troubleshooting guide

#### `scripts/setup-branch-protection.sh`
Automated script to configure branch protection using GitHub CLI.

**Protected Branches:** `main` and `develop`

**Protection Rules:**
- Require pull request before merging
- Require 1 approval
- Require all status checks to pass
- Require conversation resolution
- Dismiss stale reviews on new commits

## Workflow Triggers

### CI Workflow
```yaml
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]
```

### Docker Workflow
```yaml
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]
```

## Required Status Checks

To ensure code quality, the following checks must pass before merging:

### Lint Jobs (3)
- `lint (order-service)`
- `lint (analytics-service)`
- `lint (notification-worker)`

### Test Jobs (3)
- `test (order-service)`
- `test (analytics-service)`
- `test (notification-worker)`

### Build Jobs (3)
- `build (order-service)`
- `build (analytics-service)`
- `build (notification-worker)`

### Integration Job (1)
- `integration`

### Docker Jobs (4)
- `docker-build (order-service)`
- `docker-build (analytics-service)`
- `docker-build (notification-worker)`
- `docker-compose`

**Total: 14 checks** that must pass before merging to protected branches

## Caching Strategy

### Go Modules Cache
```yaml
~/.cache/go-build
~/go/pkg/mod
```
Key: `${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}`

### Docker BuildKit Cache
```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

## Environment Variables

### Go Version
```yaml
GO_VERSION: '1.21'
```

### Service Ports
- Order Service: `8080`
- Analytics Service: `8081`
- Notification Worker: `8082` (health check)

## Code Coverage

Code coverage is collected for all services and can be uploaded to Codecov:
```yaml
- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v4
```

To enable, add `CODECOV_TOKEN` to repository secrets.

## Local Testing

### Run Linter Locally
```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# Run for a specific service
cd services/order-service
golangci-lint run
```

### Run Tests Locally
```bash
cd services/order-service
go test -v -race -coverprofile=coverage.out ./...
```

### Build Locally
```bash
cd services/order-service
go build -o ../../bin/order-service ./cmd/order-api/main.go
```

### Test Health Checks Locally
```bash
# Start services
docker-compose up -d

# Check health
curl http://localhost:8080/health | jq
curl http://localhost:8081/health | jq
curl http://localhost:8082/health | jq
```

## Next Steps

1. **Push changes to GitHub** to trigger the workflows for the first time
2. **Set up branch protection** using the provided script or manual guide
3. **Add Codecov token** (optional) for coverage reporting
4. **Create a test PR** to verify all checks work correctly
5. **Review and adjust** linter rules based on your team's preferences

## Troubleshooting

### Workflows not running
- Check that workflows are in `.github/workflows/` directory
- Ensure YAML syntax is correct
- Check repository settings → Actions → Allow all actions

### Lint failures
- Run linter locally to see specific issues
- Adjust `.golangci.yml` if rules are too strict
- Some checks can be disabled for test files

### Health checks failing in CI
- Ensure services have enough time to start (adjust wait time)
- Check Docker Compose logs for errors
- Verify port mappings are correct

### Can't merge PR
- Ensure all status checks are passing
- Update branch with latest changes from base branch
- Resolve all conversations
- Get required approvals

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Documentation](https://golangci-lint.run/)
- [Branch Protection Documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches)
- [Docker BuildKit Cache](https://docs.docker.com/build/cache/)

---

For detailed branch protection setup, see [BRANCH_PROTECTION.md](BRANCH_PROTECTION.md)
