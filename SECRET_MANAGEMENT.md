# Secret Management with Google Cloud Secret Manager

This guide explains how to securely manage environment variables using Google Cloud Secret Manager, ensuring sensitive values like `CLERK_SECRET_KEY` are never exposed in your codebase or GitHub.

## Overview

The Golf League Manager uses Google Cloud Secret Manager to store sensitive configuration values:

- **GCP_PROJECT_ID**: Your Google Cloud project ID
- **CLERK_SECRET_KEY**: Your Clerk authentication secret key
- **CLERK_PUBLISHABLE_KEY** (optional): Your Clerk publishable key for frontend

## Initial Setup

### 1. Set Up Secrets in Google Cloud

Run the setup script to create secrets in Secret Manager:

```powershell
.\setup-secrets.ps1 `
    -ProjectId "your-gcp-project-id" `
    -ClerkSecretKey "sk_test_your_clerk_secret_key" `
    -ClerkPublishableKey "pk_test_your_clerk_publishable_key"
```

This script will:
- Enable the Secret Manager API
- Create secrets for your sensitive values
- Grant the Cloud Run service account access to these secrets
- Store the latest version of each secret

### 2. Verify Secrets Were Created

You can verify the secrets in the Google Cloud Console:

1. Go to [Secret Manager](https://console.cloud.google.com/security/secret-manager)
2. You should see:
   - `gcp-project-id`
   - `clerk-secret-key`
   - `clerk-publishable-key` (if provided)

## Local Development

### Option 1: Load Secrets from Secret Manager (Recommended)

Load secrets directly from Secret Manager into your PowerShell session:

```powershell
.\load-secrets-local.ps1 -ProjectId "your-gcp-project-id"
```

This will:
- Load all secrets from Secret Manager
- Set them as environment variables in your current PowerShell session
- Display the configuration (with masked sensitive values)

Then start the server:

```powershell
cd server\cmd
go run main.go
```

### Option 2: Manual Environment Variables

Set environment variables manually in your PowerShell session:

```powershell
$env:GCP_PROJECT_ID = "your-gcp-project-id"
$env:CLERK_SECRET_KEY = "sk_test_your_clerk_secret_key"
$env:PORT = "8080"
$env:ENVIRONMENT = "dev"
$env:LOG_LEVEL = "DEBUG"
$env:CORS_ORIGINS = "http://localhost:3000,http://localhost:5173"

cd server\cmd
go run main.go
```

### Option 3: Using .env File (Not Recommended for Production)

You can create a `.env` file in the root directory (already in `.gitignore`):

```bash
GCP_PROJECT_ID=your-gcp-project-id
CLERK_SECRET_KEY=sk_test_your_clerk_secret_key
PORT=8080
ENVIRONMENT=dev
LOG_LEVEL=DEBUG
CORS_ORIGINS=http://localhost:3000,http://localhost:5173
```

**Note**: The `.env` file is already in `.gitignore` to prevent accidental commits.

## Cloud Run Deployment

### Automatic Secret Loading

When you deploy to Cloud Run, secrets are automatically loaded from Secret Manager:

```powershell
.\deploy.ps1 `
    -ProjectId "your-gcp-project-id" `
    -ClerkPublishableKey "pk_test_..." `
    -ClerkSecretKey "sk_test_..."
```

The deployment script will:
1. Run `setup-secrets.ps1` to ensure secrets are up-to-date
2. Deploy the backend with only non-sensitive environment variables
3. The server will automatically load secrets from Secret Manager at startup

### How It Works

The server detects it's running on Cloud Run by checking for the `K_SERVICE` environment variable. When detected:

1. The server loads secrets from Secret Manager using the `secrets` package
2. Secrets are set as environment variables before configuration is loaded
3. The `config` package reads from environment variables as usual

## Updating Secrets

To update a secret value:

```powershell
# Update a single secret
echo "new-secret-value" | gcloud secrets versions add clerk-secret-key --data-file=-

# Or re-run the setup script with new values
.\setup-secrets.ps1 `
    -ProjectId "your-gcp-project-id" `
    -ClerkSecretKey "sk_test_new_value"
```

After updating secrets, redeploy your Cloud Run service to use the new values.

## Security Best Practices

### ✅ DO:
- Use Secret Manager for all sensitive values
- Keep secrets out of your codebase and version control
- Use the provided scripts to manage secrets
- Rotate secrets regularly
- Use different secrets for different environments (dev, staging, production)

### ❌ DON'T:
- Commit `.env` files to Git (already in `.gitignore`)
- Pass secrets as command-line arguments
- Log secret values
- Share secrets via email or chat
- Use production secrets in development

## Troubleshooting

### "Failed to load secrets from Secret Manager"

**Local Development**: This is normal. The server will fall back to environment variables.

**Cloud Run**: Check that:
1. Secrets exist in Secret Manager
2. The Cloud Run service account has `roles/secretmanager.secretAccessor` permission
3. The secret names match exactly: `gcp-project-id`, `clerk-secret-key`

### "GCP_PROJECT_ID environment variable is required"

**Local Development**: Run `.\load-secrets-local.ps1` or set the environment variable manually.

**Cloud Run**: Ensure the `GOOGLE_CLOUD_PROJECT` environment variable is set in the deployment.

### Permission Denied Errors

Ensure the Cloud Run service account has access:

```powershell
$projectNumber = gcloud projects describe YOUR-PROJECT-ID --format="value(projectNumber)"
$serviceAccount = "$projectNumber-compute@developer.gserviceaccount.com"

gcloud secrets add-iam-policy-binding gcp-project-id `
    --member="serviceAccount:$serviceAccount" `
    --role="roles/secretmanager.secretAccessor"
```

## Architecture

### Secret Loading Flow

```
┌─────────────────────────────────────────────────────────────┐
│                     Server Startup                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │ Check K_SERVICE │
                    │  env variable   │
                    └─────────────────┘
                              │
                ┌─────────────┴─────────────┐
                │                           │
         ┌──────▼──────┐           ┌───────▼────────┐
         │  Cloud Run  │           │     Local      │
         │  (detected) │           │  Development   │
         └──────┬──────┘           └───────┬────────┘
                │                           │
                ▼                           ▼
    ┌───────────────────────┐   ┌──────────────────────┐
    │ Load from Secret      │   │ Use Environment      │
    │ Manager API           │   │ Variables            │
    └───────────────────────┘   └──────────────────────┘
                │                           │
                └─────────────┬─────────────┘
                              ▼
                    ┌─────────────────┐
                    │ Load Config     │
                    │ from Env Vars   │
                    └─────────────────┘
```

## Files

- **`setup-secrets.ps1`**: Creates and configures secrets in Secret Manager
- **`load-secrets-local.ps1`**: Loads secrets for local development
- **`deploy.ps1`**: Deploys to Cloud Run with automatic secret setup
- **`server/internal/secrets/secrets.go`**: Go package for loading secrets
- **`.gitignore`**: Ensures `.env` files are never committed

## Additional Resources

- [Google Cloud Secret Manager Documentation](https://cloud.google.com/secret-manager/docs)
- [Clerk Authentication Documentation](https://clerk.com/docs)
- [Cloud Run Environment Variables](https://cloud.google.com/run/docs/configuring/environment-variables)
