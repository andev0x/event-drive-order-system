#!/bin/bash

# Branch Protection Setup Script
# This script sets up branch protection rules using GitHub CLI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="andev0x/event-drive-order-system"
BRANCHES=("main" "develop")

echo -e "${GREEN}Branch Protection Setup for $REPO${NC}"
echo "=========================================="

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed${NC}"
    echo "Please install it from: https://cli.github.com/"
    exit 1
fi

# Check if user is authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${YELLOW}You need to authenticate with GitHub CLI${NC}"
    gh auth login
fi

# Function to setup branch protection
setup_branch_protection() {
    local branch=$1
    echo ""
    echo -e "${YELLOW}Setting up protection for branch: $branch${NC}"
    
    # Create the protection rule
    gh api repos/$REPO/branches/$branch/protection \
      --method PUT \
      --field required_status_checks[strict]=true \
      --field 'required_status_checks[contexts][]=lint (order-service)' \
      --field 'required_status_checks[contexts][]=lint (analytics-service)' \
      --field 'required_status_checks[contexts][]=lint (notification-worker)' \
      --field 'required_status_checks[contexts][]=test (order-service)' \
      --field 'required_status_checks[contexts][]=test (analytics-service)' \
      --field 'required_status_checks[contexts][]=test (notification-worker)' \
      --field 'required_status_checks[contexts][]=build (order-service)' \
      --field 'required_status_checks[contexts][]=build (analytics-service)' \
      --field 'required_status_checks[contexts][]=build (notification-worker)' \
      --field 'required_status_checks[contexts][]=integration' \
      --field 'required_status_checks[contexts][]=docker-build (order-service)' \
      --field 'required_status_checks[contexts][]=docker-build (analytics-service)' \
      --field 'required_status_checks[contexts][]=docker-build (notification-worker)' \
      --field 'required_status_checks[contexts][]=docker-compose' \
      --field required_pull_request_reviews[dismiss_stale_reviews]=true \
      --field required_pull_request_reviews[require_code_owner_reviews]=false \
      --field required_pull_request_reviews[required_approving_review_count]=1 \
      --field required_conversation_resolution[enabled]=true \
      --field enforce_admins=false \
      --field restrictions=null \
      && echo -e "${GREEN}✓ Branch protection configured for $branch${NC}" \
      || echo -e "${RED}✗ Failed to configure branch protection for $branch${NC}"
}

# Setup protection for each branch
for branch in "${BRANCHES[@]}"; do
    # Check if branch exists
    if gh api repos/$REPO/branches/$branch &> /dev/null; then
        setup_branch_protection "$branch"
    else
        echo -e "${YELLOW}⚠ Branch '$branch' does not exist, skipping...${NC}"
    fi
done

echo ""
echo -e "${GREEN}=========================================="
echo "Branch protection setup complete!"
echo "==========================================${NC}"
echo ""
echo "To verify, run:"
echo "  gh api repos/$REPO/branches/main/protection"
echo ""
echo "Or visit:"
echo "  https://github.com/$REPO/settings/branches"
