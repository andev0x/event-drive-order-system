# Branch Protection Setup Guide

This guide will help you set up branch protection rules for your repository to ensure code quality and safety.

## Prerequisites

- Repository admin access
- GitHub CLI (`gh`) installed (optional, for automated setup)

## Method 1: GitHub Web Interface

### Steps:

1. Navigate to your repository on GitHub
2. Click on **Settings** tab
3. Click on **Branches** in the left sidebar
4. Under "Branch protection rules", click **Add rule**
5. Configure the following settings:

#### Branch name pattern
```
main
```

#### Protection rules to enable:

**Require a pull request before merging:**
- ✅ Enable
- ✅ Require approvals: 1
- ✅ Dismiss stale pull request approvals when new commits are pushed
- ✅ Require review from Code Owners (optional)

**Require status checks to pass before merging:**
- ✅ Enable
- ✅ Require branches to be up to date before merging
- Select the following status checks:
  - `lint (order-service)`
  - `lint (analytics-service)`
  - `lint (notification-worker)`
  - `test (order-service)`
  - `test (analytics-service)`
  - `test (notification-worker)`
  - `build (order-service)`
  - `build (analytics-service)`
  - `build (notification-worker)`
  - `docker-build (order-service)`
  - `docker-build (analytics-service)`
  - `docker-build (notification-worker)`
  - `docker-compose`

**Additional settings:**
- ✅ Require conversation resolution before merging
- ✅ Do not allow bypassing the above settings
- ✅ Restrict who can push to matching branches (optional)

6. Click **Create** or **Save changes**

7. **Repeat for `develop` branch** (if you use it)

---

## Method 2: GitHub CLI (Automated)

You can use the GitHub CLI to set up branch protection programmatically.

### Setup Script

Save this as `setup-branch-protection.sh` and run it:

```bash
#!/bin/bash

# Repository (format: owner/repo)
REPO="andev0x/event-drive-order-system"
BRANCH="main"

echo "Setting up branch protection for $REPO on branch $BRANCH..."

gh api repos/$REPO/branches/$BRANCH/protection \
  --method PUT \
  --field required_status_checks[strict]=true \
  --field required_status_checks[contexts][]=lint \
  --field required_status_checks[contexts][]=test \
  --field required_status_checks[contexts][]=build \
  --field required_status_checks[contexts][]=docker-build \
  --field required_status_checks[contexts][]=docker-compose \
  --field required_pull_request_reviews[dismiss_stale_reviews]=true \
  --field required_pull_request_reviews[require_code_owner_reviews]=false \
  --field required_pull_request_reviews[required_approving_review_count]=1 \
  --field required_conversation_resolution[enabled]=true \
  --field enforce_admins=true \
  --field restrictions=null

echo "Branch protection configured successfully!"
```

### Run the script:

```bash
chmod +x setup-branch-protection.sh
./setup-branch-protection.sh
```

---

## Method 3: Quick GitHub CLI Command

For a simple setup, run this single command:

```bash
gh api repos/andev0x/event-drive-order-system/branches/main/protection \
  --method PUT \
  --field required_status_checks[strict]=true \
  --field required_pull_request_reviews[required_approving_review_count]=1 \
  --field enforce_admins=false
```

---

## Verifying Branch Protection

To verify that branch protection is set up correctly:

### Using GitHub Web Interface:
1. Go to Settings → Branches
2. You should see the protection rule listed

### Using GitHub CLI:
```bash
gh api repos/andev0x/event-drive-order-system/branches/main/protection
```

### Test it:
1. Try to push directly to `main` - should be blocked
2. Create a PR and try to merge before checks pass - should be blocked
3. All CI checks must pass before merging is allowed

---

## Recommended Protection Rules Summary

| Rule | Main | Develop |
|------|------|---------|
| Require PR before merging | ✅ | ✅ |
| Require 1 approval | ✅ | ✅ |
| Require status checks | ✅ | ✅ |
| Require branch up to date | ✅ | ✅ |
| Require conversation resolution | ✅ | ✅ |
| Enforce for admins | ✅ | ⚠️ |

---

## Status Checks to Require

After your first PR with GitHub Actions, these checks will appear and should be marked as required:

- **CI Workflow:**
  - `lint (order-service)`
  - `lint (analytics-service)`
  - `lint (notification-worker)`
  - `test (order-service)`
  - `test (analytics-service)`
  - `test (notification-worker)`
  - `build (order-service)`
  - `build (analytics-service)`
  - `build (notification-worker)`
  - `integration`

- **Docker Workflow:**
  - `docker-build (order-service)`
  - `docker-build (analytics-service)`
  - `docker-build (notification-worker)`
  - `docker-compose`

---

## Troubleshooting

### Status checks don't appear
- Push a commit to trigger the GitHub Actions workflows
- Once they run, the checks will appear in the branch protection settings

### Can't merge even with passing checks
- Ensure branch is up to date with base branch
- Check that all conversations are resolved

### Accidentally locked out
- If you set "Enforce for admins" and can't merge, temporarily disable it in settings

---

## Next Steps

1. ✅ Push this configuration to your repository
2. ✅ Set up branch protection rules
3. ✅ Create a test PR to verify everything works
4. ✅ Ensure all team members understand the workflow

For more information, see the [GitHub Branch Protection Documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches).
