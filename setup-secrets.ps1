# Golf League Manager - Secret Manager Setup Script
# This script creates secrets in Google Cloud Secret Manager

param(
    [Parameter(Mandatory = $true)]
    [string]$ProjectId,
    
    [Parameter(Mandatory = $true)]
    [string]$ClerkSecretKey,
    
    [Parameter(Mandatory = $false)]
    [string]$ClerkPublishableKey = ""
)

Write-Host "🔐 Setting up secrets in Google Cloud Secret Manager..." -ForegroundColor Green
Write-Host "Project ID: $ProjectId" -ForegroundColor Cyan

# Set the GCP project
Write-Host "`n📦 Setting GCP project..." -ForegroundColor Yellow
gcloud config set project $ProjectId

# Enable Secret Manager API if not already enabled
Write-Host "`n🔧 Enabling Secret Manager API..." -ForegroundColor Yellow
gcloud services enable secretmanager.googleapis.com

# Create or update GCP_PROJECT_ID secret
Write-Host "`n📝 Creating/updating GCP_PROJECT_ID secret..." -ForegroundColor Yellow
$projectIdExists = gcloud secrets describe gcp-project-id 2>$null
if ($LASTEXITCODE -eq 0) {
    Write-Host "Secret 'gcp-project-id' already exists, adding new version..." -ForegroundColor Cyan
    echo $ProjectId | gcloud secrets versions add gcp-project-id --data-file=-
} else {
    Write-Host "Creating new secret 'gcp-project-id'..." -ForegroundColor Cyan
    echo $ProjectId | gcloud secrets create gcp-project-id --data-file=-
}

# Create or update CLERK_SECRET_KEY secret
Write-Host "`n📝 Creating/updating CLERK_SECRET_KEY secret..." -ForegroundColor Yellow
$clerkSecretExists = gcloud secrets describe clerk-secret-key 2>$null
if ($LASTEXITCODE -eq 0) {
    Write-Host "Secret 'clerk-secret-key' already exists, adding new version..." -ForegroundColor Cyan
    echo $ClerkSecretKey | gcloud secrets versions add clerk-secret-key --data-file=-
} else {
    Write-Host "Creating new secret 'clerk-secret-key'..." -ForegroundColor Cyan
    echo $ClerkSecretKey | gcloud secrets create clerk-secret-key --data-file=-
}

# Optionally create CLERK_PUBLISHABLE_KEY secret
if ($ClerkPublishableKey -ne "") {
    Write-Host "`n📝 Creating/updating CLERK_PUBLISHABLE_KEY secret..." -ForegroundColor Yellow
    $clerkPubExists = gcloud secrets describe clerk-publishable-key 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Secret 'clerk-publishable-key' already exists, adding new version..." -ForegroundColor Cyan
        echo $ClerkPublishableKey | gcloud secrets versions add clerk-publishable-key --data-file=-
    } else {
        Write-Host "Creating new secret 'clerk-publishable-key'..." -ForegroundColor Cyan
        echo $ClerkPublishableKey | gcloud secrets create clerk-publishable-key --data-file=-
    }
}

# Grant Cloud Run service account access to secrets
Write-Host "`n🔑 Granting Cloud Run service account access to secrets..." -ForegroundColor Yellow
$projectNumber = gcloud projects describe $ProjectId --format="value(projectNumber)"
$serviceAccount = "$projectNumber-compute@developer.gserviceaccount.com"

Write-Host "Service Account: $serviceAccount" -ForegroundColor Cyan

gcloud secrets add-iam-policy-binding gcp-project-id `
    --member="serviceAccount:$serviceAccount" `
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding clerk-secret-key `
    --member="serviceAccount:$serviceAccount" `
    --role="roles/secretmanager.secretAccessor"

if ($ClerkPublishableKey -ne "") {
    gcloud secrets add-iam-policy-binding clerk-publishable-key `
        --member="serviceAccount:$serviceAccount" `
        --role="roles/secretmanager.secretAccessor"
}

Write-Host "`n✅ Secret Manager setup complete!" -ForegroundColor Green
Write-Host "`n📊 Summary:" -ForegroundColor Cyan
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host "✓ Secret 'gcp-project-id' configured" -ForegroundColor White
Write-Host "✓ Secret 'clerk-secret-key' configured" -ForegroundColor White
if ($ClerkPublishableKey -ne "") {
    Write-Host "✓ Secret 'clerk-publishable-key' configured" -ForegroundColor White
}
Write-Host "✓ Cloud Run service account granted access" -ForegroundColor White
Write-Host "===========================================" -ForegroundColor Cyan

Write-Host "`n💡 Next steps:" -ForegroundColor Yellow
Write-Host "1. For local development, run: .\load-secrets-local.ps1 -ProjectId $ProjectId" -ForegroundColor White
Write-Host "2. Deploy to Cloud Run with: .\deploy.ps1 (secrets will be loaded automatically)" -ForegroundColor White
