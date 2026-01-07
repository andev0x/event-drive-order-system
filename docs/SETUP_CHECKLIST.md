# GitHub Actions Setup Checklist

Use this checklist to ensure everything is properly configured.

## ✅ Files Created

- [ ] `.github/workflows/ci.yml` - Main CI workflow
- [ ] `.github/workflows/docker.yml` - Docker build workflow
- [ ] `.golangci.yml` - Linter configuration
- [ ] `docs/GITHUB_ACTIONS_SETUP.md` - Setup documentation
- [ ] `docs/BRANCH_PROTECTION.md` - Branch protection guide
- [ ] `scripts/setup-branch-protection.sh` - Automation script

## ✅ Code Changes

### Order Service
- [ ] Enhanced health check in `internal/handler/order_handler.go`
- [ ] Health checker setup in `cmd/order-api/main.go`
- [ ] HealthCheck method in `internal/mq/publisher.go`

### Analytics Service
- [ ] Enhanced health check in `internal/handler/analytics_handler.go`
- [ ] Health checker setup in `cmd/analytics-api/main.go`
- [ ] HealthCheck method in `internal/mq/consumer.go`

### Notification Worker
- [ ] HTTP health check server in `cmd/notification-worker/main.go`
- [ ] Health endpoint on port 8082

## ✅ Pre-Push Verification

### 1. Test Builds Locally
```bash
cd services/order-service && go build ./cmd/order-api/main.go
cd ../analytics-service && go build ./cmd/analytics-api/main.go
cd ../notification-worker && go build ./cmd/notification-worker/main.go
```
All should build without errors.

### 2. Test Health Endpoints Locally (Optional)
```bash
# Start services
docker-compose up -d

# Wait for services to start
sleep 30

# Test health checks
curl http://localhost:8080/health | jq
curl http://localhost:8081/health | jq
curl http://localhost:8082/health | jq

# Cleanup
docker-compose down
```

### 3. Test Linter (Optional)
```bash
# Install golangci-lint if not already installed
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# Run linter
cd services/order-service
golangci-lint run --timeout=5m
```

## ✅ After Pushing to GitHub

### 1. Verify Workflows Run
- [ ] Go to repository → Actions tab
- [ ] Check that both CI and Docker workflows are running
- [ ] All jobs should complete successfully (may have some lint warnings initially)

### 2. Review Workflow Results
- [ ] Lint jobs completed
- [ ] Test jobs passed
- [ ] Build jobs succeeded
- [ ] Integration job passed
- [ ] Docker build jobs succeeded
- [ ] Docker compose validation passed

### 3. Set Up Branch Protection

#### Option A: Use the Script
```bash
# Make sure you have GitHub CLI installed and authenticated
gh auth login

# Run the setup script
./scripts/setup-branch-protection.sh
```

#### Option B: Manual Setup
Follow the guide in `docs/BRANCH_PROTECTION.md`

### 4. Enable Required Status Checks
- [ ] Go to Settings → Branches
- [ ] Edit the branch protection rule for `main`
- [ ] Under "Require status checks to pass before merging", select:
  - [ ] `lint (order-service)`
  - [ ] `lint (analytics-service)`
  - [ ] `lint (notification-worker)`
  - [ ] `test (order-service)`
  - [ ] `test (analytics-service)`
  - [ ] `test (notification-worker)`
  - [ ] `build (order-service)`
  - [ ] `build (analytics-service)`
  - [ ] `build (notification-worker)`
  - [ ] `integration`
  - [ ] `docker-build (order-service)`
  - [ ] `docker-build (analytics-service)`
  - [ ] `docker-build (notification-worker)`
  - [ ] `docker-compose`

## ✅ Test the Setup

### Create a Test Pull Request
```bash
# Create a new branch
git checkout -b test-ci-setup

# Make a small change
echo "# CI/CD Test" >> docs/CI_TEST.md

# Commit and push
git add docs/CI_TEST.md
git commit -m "test: verify CI/CD pipeline"
git push origin test-ci-setup

# Create PR using GitHub CLI
gh pr create --title "Test CI/CD Pipeline" --body "Testing the new CI/CD setup"
```

### Verify PR Checks
- [ ] All CI checks appear in the PR
- [ ] All checks are passing
- [ ] Cannot merge until checks pass (if branch protection is enabled)
- [ ] Can merge after checks pass

### Clean Up Test PR
```bash
# After verification, close and delete
gh pr close test-ci-setup
git checkout main
git branch -D test-ci-setup
```

## ✅ Optional Enhancements

### 1. Code Coverage (Codecov)
- [ ] Sign up at [codecov.io](https://codecov.io)
- [ ] Get repository token
- [ ] Add `CODECOV_TOKEN` to repository secrets
- [ ] Coverage reports will be automatically uploaded

### 2. Additional Workflows
Consider adding:
- [ ] Release workflow for tagged versions
- [ ] Deployment workflow for staging/production
- [ ] Dependency update workflow (Dependabot)
- [ ] Security scanning workflow

### 3. Pre-commit Hooks
Add local pre-commit hooks to catch issues before pushing:
```bash
# Create .git/hooks/pre-commit
#!/bin/bash
golangci-lint run --timeout=5m
go test ./...
```

## 🎉 Success Criteria

Your setup is complete when:
- ✅ All workflows run successfully on push
- ✅ Branch protection is active on `main` and `develop`
- ✅ Health checks return 200 OK for all services
- ✅ Linter runs without errors
- ✅ All tests pass
- ✅ Docker images build successfully
- ✅ Test PR shows all required checks

## 📚 Documentation

- Main guide: `docs/GITHUB_ACTIONS_SETUP.md`
- Branch protection: `docs/BRANCH_PROTECTION.md`
- Project README: `README.md`

## 🆘 Troubleshooting

If something doesn't work:
1. Check workflow logs in GitHub Actions tab
2. Review the troubleshooting section in `docs/GITHUB_ACTIONS_SETUP.md`
3. Verify all files are committed and pushed
4. Ensure workflows have correct YAML syntax
5. Check repository Actions settings are not restricted

---

**Need help?** Check the documentation files or create an issue in the repository.
