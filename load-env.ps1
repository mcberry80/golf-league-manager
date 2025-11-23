# Golf League Manager - Load Environment Variables
# This script loads CLERK_SECRET_KEY from Secret Manager and sets all environment variables

Write-Host "🔐 Loading environment variables..." -ForegroundColor Green

# Set GCP Project ID (not sensitive)
$env:GCP_PROJECT_ID = "elite-league-manager"

# Set development defaults
$env:PORT = "8080"
$env:ENVIRONMENT = "dev"
$env:LOG_LEVEL = "DEBUG"
$env:CORS_ORIGINS = "http://localhost:3000,http://localhost:5173"

# Load CLERK_SECRET_KEY from Secret Manager
Write-Host "`n📥 Loading CLERK_SECRET_KEY from Secret Manager..." -ForegroundColor Yellow
$clerkSecretKey = gcloud secrets versions access latest --secret="clerk-secret-key" --project="$env:GCP_PROJECT_ID" 2>$null
if ($LASTEXITCODE -eq 0) {
    $env:CLERK_SECRET_KEY = $clerkSecretKey.Trim()
    Write-Host "✓ CLERK_SECRET_KEY loaded" -ForegroundColor Green
}
else {
    Write-Host "✗ Failed to load CLERK_SECRET_KEY" -ForegroundColor Red
    Write-Host "`nPlease set it manually or check Secret Manager configuration" -ForegroundColor Yellow
}

Write-Host "`n✅ Environment variables loaded!" -ForegroundColor Green
Write-Host "`n📊 Current configuration:" -ForegroundColor Cyan
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host "GCP_PROJECT_ID: $env:GCP_PROJECT_ID" -ForegroundColor White
if ($env:CLERK_SECRET_KEY) {
    Write-Host "CLERK_SECRET_KEY: $($env:CLERK_SECRET_KEY.Substring(0, [Math]::Min(8, $env:CLERK_SECRET_KEY.Length)))****" -ForegroundColor White
}
else {
    Write-Host "CLERK_SECRET_KEY: NOT SET" -ForegroundColor Red
}
Write-Host "PORT: $env:PORT" -ForegroundColor White
Write-Host "ENVIRONMENT: $env:ENVIRONMENT" -ForegroundColor White
Write-Host "LOG_LEVEL: $env:LOG_LEVEL" -ForegroundColor White
Write-Host "CORS_ORIGINS: $env:CORS_ORIGINS" -ForegroundColor White
Write-Host "===========================================" -ForegroundColor Cyan

Write-Host "`n💡 To start the server, run:" -ForegroundColor Yellow
Write-Host "cd server\cmd" -ForegroundColor White
Write-Host "go run main.go" -ForegroundColor White

Write-Host "`n⚠ Note: These environment variables are only set for this PowerShell session." -ForegroundColor Yellow
