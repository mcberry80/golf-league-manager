# Golf League Manager - Load Secrets for Local Development
# This script loads secrets from Google Cloud Secret Manager and sets them as environment variables

param(
    [Parameter(Mandatory = $true)]
    [string]$ProjectId
)

Write-Host "🔐 Loading secrets from Google Cloud Secret Manager..." -ForegroundColor Green
Write-Host "Project ID: $ProjectId" -ForegroundColor Cyan

# Set the GCP project
gcloud config set project $ProjectId 2>$null

# Load GCP_PROJECT_ID
Write-Host "`n📥 Loading GCP_PROJECT_ID..." -ForegroundColor Yellow
$gcpProjectId = gcloud secrets versions access latest --secret="gcp-project-id" 2>$null
if ($LASTEXITCODE -eq 0) {
    $env:GCP_PROJECT_ID = $gcpProjectId.Trim()
    Write-Host "✓ GCP_PROJECT_ID loaded" -ForegroundColor Green
}
else {
    Write-Host "✗ Failed to load GCP_PROJECT_ID" -ForegroundColor Red
}

# Load CLERK_SECRET_KEY
Write-Host "`n📥 Loading CLERK_SECRET_KEY..." -ForegroundColor Yellow
$clerkSecretKey = gcloud secrets versions access latest --secret="clerk-secret-key" 2>$null
if ($LASTEXITCODE -eq 0) {
    $env:CLERK_SECRET_KEY = $clerkSecretKey.Trim()
    Write-Host "✓ CLERK_SECRET_KEY loaded" -ForegroundColor Green
}
else {
    Write-Host "✗ Failed to load CLERK_SECRET_KEY" -ForegroundColor Red
}

# Load CLERK_PUBLISHABLE_KEY (optional)
Write-Host "`n📥 Loading CLERK_PUBLISHABLE_KEY (optional)..." -ForegroundColor Yellow
$clerkPubKey = gcloud secrets versions access latest --secret="clerk-publishable-key" 2>$null
if ($LASTEXITCODE -eq 0) {
    $env:CLERK_PUBLISHABLE_KEY = $clerkPubKey.Trim()
    Write-Host "✓ CLERK_PUBLISHABLE_KEY loaded" -ForegroundColor Green
}
else {
    Write-Host "⚠ CLERK_PUBLISHABLE_KEY not found (optional)" -ForegroundColor Yellow
}

# Set default values for other environment variables
$env:PORT = "8080"
$env:ENVIRONMENT = "dev"
$env:LOG_LEVEL = "DEBUG"
$env:CORS_ORIGINS = "http://localhost:3000,http://localhost:5173"

Write-Host "`n✅ Environment variables loaded!" -ForegroundColor Green
Write-Host "`n📊 Current configuration:" -ForegroundColor Cyan
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host "GCP_PROJECT_ID: $env:GCP_PROJECT_ID" -ForegroundColor White
Write-Host "CLERK_SECRET_KEY: $($env:CLERK_SECRET_KEY.Substring(0, [Math]::Min(8, $env:CLERK_SECRET_KEY.Length)))****" -ForegroundColor White
Write-Host "PORT: $env:PORT" -ForegroundColor White
Write-Host "ENVIRONMENT: $env:ENVIRONMENT" -ForegroundColor White
Write-Host "LOG_LEVEL: $env:LOG_LEVEL" -ForegroundColor White
Write-Host "CORS_ORIGINS: $env:CORS_ORIGINS" -ForegroundColor White
Write-Host "===========================================" -ForegroundColor Cyan

Write-Host "`n💡 To start the server, run:" -ForegroundColor Yellow
Write-Host "cd server\cmd" -ForegroundColor White
Write-Host "go run main.go" -ForegroundColor White

Write-Host "`n⚠ Note: These environment variables are only set for this PowerShell session." -ForegroundColor Yellow
Write-Host "To persist them, add them to your system environment variables or use a .env file." -ForegroundColor Yellow
