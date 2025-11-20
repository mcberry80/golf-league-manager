# Golf League Manager - Deployment Guide

## Quick Deploy to GCP

Run the deployment script with your credentials:

```powershell
.\deploy.ps1 `
    -ProjectId "your-gcp-project-id" `
    -ClerkPublishableKey "pk_test_..." `
    -ClerkSecretKey "sk_test_..." `
    -Region "us-central1"
```

## Prerequisites

1. **GCP Setup**:
   - Google Cloud account with billing enabled
   - `gcloud` CLI installed and authenticated (`gcloud auth login`)
   - Firestore database created in your project
   - Cloud Run API enabled
   - Cloud Build API enabled

2. **Clerk Setup**:
   - Clerk account created at [clerk.com](https://clerk.com)
   - Application created in Clerk dashboard
   - Get your Publishable Key and Secret Key

## Manual Deployment

### Backend

```powershell
cd server
gcloud builds submit --tag gcr.io/YOUR-PROJECT-ID/golf-league-backend
gcloud run deploy golf-league-backend `
    --image gcr.io/YOUR-PROJECT-ID/golf-league-backend `
    --platform managed `
    --region us-central1 `
    --allow-unauthenticated `
    --set-env-vars "GCP_PROJECT_ID=YOUR-PROJECT-ID,CLERK_SECRET_KEY=sk_test_..."
```

### Frontend

```powershell
cd frontend

# Update .env.production with backend URL
echo "VITE_CLERK_PUBLISHABLE_KEY=pk_test_..." > .env.production
echo "VITE_API_URL=https://your-backend-url.run.app" >> .env.production

gcloud builds submit --tag gcr.io/YOUR-PROJECT-ID/golf-league-frontend
gcloud run deploy golf-league-frontend `
    --image gcr.io/YOUR-PROJECT-ID/golf-league-frontend `
    --platform managed `
    --region us-central1 `
    --allow-unauthenticated
```

## Post-Deployment

1. **Update Clerk Settings**:
   - Go to Clerk Dashboard → Your App → Paths
   - Add your frontend URL to allowed redirect URLs
   - Add your frontend URL to allowed origins

2. **Test the Application**:
   - Visit the frontend URL
   - Sign in with Clerk
   - Verify all pages load without 404 errors

3. **Create Admin User**:
   - In Firestore console, find your player document
   - Set `is_admin: true` for the admin user

## Local Development

### Backend
```powershell
cd server\cmd
$env:GCP_PROJECT_ID="your-project-id"
$env:CLERK_SECRET_KEY="sk_test_..."
go run main.go
```

### Frontend
```powershell
cd frontend

# Create .env.local
echo "VITE_CLERK_PUBLISHABLE_KEY=pk_test_..." > .env.local
echo "VITE_API_URL=http://localhost:8080" >> .env.local

npm install
npm run dev
```

## Troubleshooting

**404 Errors**: All routes are now configured. If you still see 404s, check browser console for errors.

**CORS Issues**: Backend is configured to allow all origins. If issues persist, check Cloud Run logs.

**Auth Issues**: Verify Clerk keys are correct and frontend URL is in Clerk's allowed origins.

**Firestore Errors**: Ensure Firestore is enabled and the service account has proper permissions.
