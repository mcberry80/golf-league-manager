# Golf League Manager - Quick Start for Local Development
# This script sets up all environment variables and starts the server

Write-Host "🏌️ Golf League Manager - Quick Start" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green

# Set GCP Project ID (not sensitive)
$env:GCP_PROJECT_ID = "elite-league-manager"

# Set development defaults
$env:PORT = "8080"
$env:ENVIRONMENT = "dev"
$env:LOG_LEVEL = "DEBUG"
$env:CORS_ORIGINS = "http://localhost:3000,http://localhost:5173"

Write-Host "`n📋 Configuration:" -ForegroundColor Cyan
Write-Host "  Project ID: $env:GCP_PROJECT_ID" -ForegroundColor White
Write-Host "  Port: $env:PORT" -ForegroundColor White
Write-Host "  Environment: $env:ENVIRONMENT" -ForegroundColor White
Write-Host "  Log Level: $env:LOG_LEVEL" -ForegroundColor White

# Load CLERK_SECRET_KEY from Secret Manager
Write-Host "`n🔐 Loading CLERK_SECRET_KEY from Secret Manager..." -ForegroundColor Yellow
try {
    $clerkSecretKey = gcloud secrets versions access latest --secret="clerk-secret-key" --project="$env:GCP_PROJECT_ID" 2>$null
    if ($LASTEXITCODE -eq 0) {
        $env:CLERK_SECRET_KEY = $clerkSecretKey.Trim()
        Write-Host "✓ CLERK_SECRET_KEY loaded successfully" -ForegroundColor Green
    }
    else {
        throw "Failed to load secret"
    }
}
catch {
    Write-Host "✗ Failed to load CLERK_SECRET_KEY from Secret Manager" -ForegroundColor Red
    Write-Host "`nPlease set it manually:" -ForegroundColor Yellow
    Write-Host "  `$env:CLERK_SECRET_KEY = 'sk_test_...'" -ForegroundColor White
    Write-Host "`nOr create the secret in Secret Manager:" -ForegroundColor Yellow
    Write-Host "  echo 'sk_test_...' | gcloud secrets create clerk-secret-key --data-file=- --project=elite-league-manager" -ForegroundColor White
    exit 1
}

Write-Host "`n✅ All environment variables configured!" -ForegroundColor Green

# Start the server
Write-Host "`n🚀 Starting backend server..." -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green
Write-Host "Server will be available at: http://localhost:$env:PORT" -ForegroundColor Cyan
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Yellow
Write-Host "=====================================" -ForegroundColor Green
Write-Host ""

Push-Location server\cmd
go run main.go
Pop-Location
