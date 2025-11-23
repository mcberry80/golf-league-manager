# Deploy Firestore Indexes
# This script deploys the Firestore composite indexes defined in firestore.indexes.json

Write-Host "Deploying Firestore indexes..." -ForegroundColor Cyan

# Check if Firebase CLI is installed
if (!(Get-Command firebase -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Firebase CLI is not installed." -ForegroundColor Red
    Write-Host "Install it with: npm install -g firebase-tools" -ForegroundColor Yellow
    exit 1
}

# Deploy indexes
firebase deploy --only firestore:indexes --project elite-league-manager

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nIndexes deployed successfully!" -ForegroundColor Green
    Write-Host "Note: It may take a few minutes for the indexes to build." -ForegroundColor Yellow
} else {
    Write-Host "`nFailed to deploy indexes." -ForegroundColor Red
    exit 1
}
