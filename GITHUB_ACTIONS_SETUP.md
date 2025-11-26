# GitHub Actions Setup Guide

This guide explains how to set up GitHub Actions for automated CI/CD deployment.

## Overview

The repository now includes 4 GitHub Actions workflows:

1. **Backend CI/CD** (`.github/workflows/backend-deploy.yml`)
   - Runs tests on every push to backend code
   - Deploys to Cloud Run on push to `main`

2. **Frontend CI/CD** (`.github/workflows/frontend-deploy.yml`)
   - Builds and deploys frontend to Cloud Run on push to `main`

3. **Firestore Indexes** (`.github/workflows/deploy-indexes.yml`)
   - Deploys Firestore indexes when `firestore.indexes.json` changes

4. **PR Validation** (`.github/workflows/pr-validation.yml`)
   - Runs tests and builds on all pull requests

## Required GitHub Secrets

Navigate to your GitHub repository → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**

Add the following secrets:

### 1. `GCP_PROJECT_ID`
Your Google Cloud Project ID (e.g., `gelite-league-manager`)

### 2. `GCP_SA_KEY`
Service account JSON key with the following permissions:
- Cloud Run Admin
- Cloud Build Editor
- Secret Manager Secret Accessor
- Storage Admin
- Service Account User

**To create the service account:**

```bash
# Set your project ID
PROJECT_ID="elite-league-manager"

# Create service account
gcloud iam service-accounts create github-actions `
  --display-name="GitHub Actions" `
  --project=$PROJECT_ID

# Grant necessary roles
gcloud projects add-iam-policy-binding $PROJECT_ID `
  --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" `
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID `
  --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" `
  --role="roles/cloudbuild.builds.editor"

gcloud projects add-iam-policy-binding $PROJECT_ID `
  --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" `
  --role="roles/secretmanager.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID `
  --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" `
  --role="roles/storage.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID `
  --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" `
  --role="roles/iam.serviceAccountUser"

# Create and download key
gcloud iam service-accounts keys create github-actions-key.json `
  --iam-account=github-actions@$PROJECT_ID.iam.gserviceaccount.com

# Copy the contents of github-actions-key.json to GitHub Secret GCP_SA_KEY
cat github-actions-key.json

# IMPORTANT: Delete the local key file after copying to GitHub
rm github-actions-key.json
```

### 3. `CLERK_SECRET_KEY`
Your Clerk secret key (starts with `sk_test_` or `sk_live_`)

Get from: https://dashboard.clerk.com → Your App → API Keys

### 4. `CLERK_PUBLISHABLE_KEY`
Your Clerk publishable key (starts with `pk_test_` or `pk_live_`)

Get from: https://dashboard.clerk.com → Your App → API Keys

## Testing the Workflows

### Test PR Validation
1. Create a new branch: `git checkout -b test-ci`
2. Make a small change to any file
3. Push and create a PR
4. The PR validation workflow should run automatically

### Test Deployment
1. Merge a PR to `main` or push directly to `main`
2. The backend and/or frontend workflows will run based on which files changed
3. Check the **Actions** tab in GitHub to see progress

### Manual Deployment
You can manually trigger any workflow:
1. Go to **Actions** tab
2. Select the workflow (e.g., "Backend CI/CD")
3. Click **Run workflow** → Select branch → **Run workflow**

## Workflow Triggers

| Workflow | Automatic Trigger | Manual Trigger |
|----------|------------------|----------------|
| Backend CI/CD | Push to `main` (backend files) | ✅ Yes |
| Frontend CI/CD | Push to `main` (frontend files) | ✅ Yes |
| Deploy Indexes | Push to `main` (firestore.indexes.json) | ✅ Yes |
| PR Validation | Any pull request | ❌ No |

## Monitoring Deployments

- **GitHub Actions Tab**: See all workflow runs, logs, and status
- **Cloud Run Console**: View deployed services and logs
- **Deployment Summary**: Each successful deployment adds a summary to the workflow run

## Local Development

The PowerShell scripts are still available for local development:

- `.\start-dev.ps1` - Start local development servers
- `.\load-secrets-local.ps1` - Load secrets from Secret Manager
- `.\setup-secrets.ps1` - Update secrets in Secret Manager

## Troubleshooting

### Workflow fails with "Permission denied"
- Check that `GCP_SA_KEY` is correctly set in GitHub Secrets
- Verify service account has all required roles

### Backend/Frontend not deploying
- Check the **Actions** tab for error messages
- Verify `GCP_PROJECT_ID` matches your actual project
- Ensure Cloud Run API is enabled in your GCP project

### Secrets not found in Secret Manager
- The workflows automatically create secrets on first run
- If issues persist, manually create secrets using `.\setup-secrets.ps1`

## Security Best Practices

✅ **DO**:
- Use separate service accounts for different environments
- Rotate service account keys periodically
- Use branch protection rules to require PR reviews

❌ **DON'T**:
- Commit service account keys to the repository
- Share GitHub Secret values
- Use production keys in development

## Next Steps

1. Set up the required GitHub Secrets
2. Create a test PR to validate the workflows
3. Merge to `main` to trigger your first automated deployment
4. Update Clerk settings with your deployed frontend URL
