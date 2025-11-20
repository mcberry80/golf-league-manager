$env:GCP_PROJECT_ID = "elite-league-manager"
$env:NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY = "pk_test_bGVnYWwtc2FpbGZpc2gtODQuY2xlcmsuYWNjb3VudHMuZGV2JA"
$env:CLERK_SECRET_KEY = "sk_test_3QOVlrVjimzYbKsLJZ1VCGrfTqeOuVuJtvE3vaQMD0"

# gcloud run deploy golf-league-backend `
#   --source ./server `
#   --region us-central1 `
#   --allow-unauthenticated `
#   --set-env-vars "GCP_PROJECT_ID=$env:GCP_PROJECT_ID,NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=$env:NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY,CLERK_SECRET_KEY=$env:CLERK_SECRET_KEY" `
#   --quiet

# Deploy Frontend
$BACKEND_URL = "https://golf-league-backend-130681562264.us-central1.run.app"

gcloud run deploy golf-league-frontend `
  --source ./frontend `
  --region us-central1 `
  --allow-unauthenticated `
  --set-env-vars "VITE_API_URL=$BACKEND_URL,VITE_CLERK_PUBLISHABLE_KEY=$env:NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY" `
  --port 80 `
  --quiet
