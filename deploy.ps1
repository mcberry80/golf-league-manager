# Golf League Manager - Deployment Script for GCP
# This script deploys both backend and frontend to Google Cloud Run

param(
  [Parameter(Mandatory = $true)]
  [string]$ProjectId,
    
  [Parameter(Mandatory = $true)]
  [string]$ClerkPublishableKey,
    
  [Parameter(Mandatory = $true)]
  [string]$ClerkSecretKey,
    
  [string]$Region = "us-central1"
)

Write-Host "🚀 Starting deployment to GCP..." -ForegroundColor Green
Write-Host "Project ID: $ProjectId" -ForegroundColor Cyan
Write-Host "Region: $Region" -ForegroundColor Cyan

# Set the GCP project
Write-Host "`n📦 Setting GCP project..." -ForegroundColor Yellow
gcloud config set project $ProjectId

# Deploy Backend
Write-Host "`n🔧 Building and deploying backend..." -ForegroundColor Yellow

# First, ensure secrets are set up in Secret Manager
Write-Host "`n🔐 Ensuring secrets are configured in Secret Manager..." -ForegroundColor Yellow
& "$PSScriptRoot\setup-secrets.ps1" -ProjectId $ProjectId -ClerkSecretKey $ClerkSecretKey -ClerkPublishableKey $ClerkPublishableKey

Push-Location server

# Build and deploy backend to Cloud Run
gcloud builds submit --tag gcr.io/$ProjectId/golf-league-backend

# Deploy with only non-sensitive environment variables
# Secrets will be loaded from Secret Manager at runtime
gcloud run deploy golf-league-backend `
  --image gcr.io/$ProjectId/golf-league-backend `
  --platform managed `
  --region $Region `
  --allow-unauthenticated `
  --set-env-vars "ENVIRONMENT=production,GOOGLE_CLOUD_PROJECT=$ProjectId"

Pop-Location


# Get backend URL
Write-Host "`n🔍 Getting backend URL..." -ForegroundColor Yellow
$backendUrl = gcloud run services describe golf-league-backend --region $Region --format="value(status.url)"
Write-Host "Backend URL: $backendUrl" -ForegroundColor Green

# Update frontend environment variables
Write-Host "`n📝 Updating frontend environment variables..." -ForegroundColor Yellow
Push-Location frontend

# Create production env file
@"
VITE_CLERK_PUBLISHABLE_KEY=$ClerkPublishableKey
VITE_API_URL=$backendUrl
"@ | Out-File -FilePath .env.production -Encoding UTF8

# Build and deploy frontend to Cloud Run
Write-Host "`n🎨 Building and deploying frontend..." -ForegroundColor Yellow
gcloud builds submit --tag gcr.io/$ProjectId/golf-league-frontend

gcloud run deploy golf-league-frontend `
  --image gcr.io/$ProjectId/golf-league-frontend `
  --platform managed `
  --region $Region `
  --allow-unauthenticated

Pop-Location

# Get frontend URL
Write-Host "`n🔍 Getting frontend URL..." -ForegroundColor Yellow
$frontendUrl = gcloud run services describe golf-league-frontend --region $Region --format="value(status.url)"

# Display results
Write-Host "`n✅ Deployment complete!" -ForegroundColor Green
Write-Host "`n📊 Deployment Summary:" -ForegroundColor Cyan
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host "Backend URL:  $backendUrl" -ForegroundColor White
Write-Host "Frontend URL: $frontendUrl" -ForegroundColor White
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host "`n🌐 Open the frontend URL in your browser to access the application." -ForegroundColor Green
Write-Host "🔐 Make sure to update your Clerk settings with the frontend URL for redirects." -ForegroundColor Yellow
