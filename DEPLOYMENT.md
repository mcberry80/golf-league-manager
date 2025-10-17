# Deployment Guide

This guide covers deploying the Golf League Manager application to various platforms.

## Table of Contents

- [Local Development with Docker](#local-development-with-docker)
- [Google Cloud Run Deployment](#google-cloud-run-deployment)
- [Vercel Frontend Deployment](#vercel-frontend-deployment)
- [Environment Configuration](#environment-configuration)

## Local Development with Docker

### Prerequisites

- Docker and Docker Compose installed
- Clerk account with API keys
- GCP Project ID (or use emulator)

### Steps

1. Create a `.env` file in the root directory:

```bash
GCP_PROJECT_ID=your-project-id
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...
```

2. Start all services:

```bash
docker-compose up
```

This will start:
- Backend API on http://localhost:8080
- Frontend on http://localhost:3000
- Firestore emulator on http://localhost:8081

## Google Cloud Run Deployment

### Prerequisites

- Google Cloud account with billing enabled
- `gcloud` CLI installed and authenticated
- Firestore database created in your project

### Backend Deployment

1. Build and push container:

```bash
# Set your project ID
export PROJECT_ID=your-project-id

# Build the backend
gcloud builds submit --tag gcr.io/$PROJECT_ID/golf-league-backend .

# Deploy to Cloud Run
gcloud run deploy golf-league-backend \
  --image gcr.io/$PROJECT_ID/golf-league-backend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars GCP_PROJECT_ID=$PROJECT_ID
```

2. Note the service URL (e.g., https://golf-league-backend-xxx-uc.a.run.app)

### Frontend Deployment Options

#### Option 1: Deploy to Cloud Run

```bash
cd frontend

# Build and push
gcloud builds submit --tag gcr.io/$PROJECT_ID/golf-league-frontend .

# Deploy
gcloud run deploy golf-league-frontend \
  --image gcr.io/$PROJECT_ID/golf-league-frontend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars NEXT_PUBLIC_API_URL=https://your-backend-url.run.app \
  --set-env-vars NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_... \
  --set-env-vars CLERK_SECRET_KEY=sk_...
```

#### Option 2: Deploy to Vercel (Recommended)

See [Vercel Frontend Deployment](#vercel-frontend-deployment) below.

### Cloud Scheduler Setup

Set up weekly handicap recalculation:

```bash
# Create scheduler job to run every Monday at 2 AM
gcloud scheduler jobs create http handicap-recalc \
  --schedule="0 2 * * 1" \
  --uri="https://your-backend-url.run.app/api/jobs/recalculate-handicaps" \
  --http-method=POST \
  --time-zone="America/New_York" \
  --location=us-central1
```

## Vercel Frontend Deployment

### Prerequisites

- Vercel account (free tier available)
- Vercel CLI (`npm i -g vercel`)

### Steps

1. Navigate to frontend directory:

```bash
cd frontend
```

2. Configure environment variables in Vercel:

Go to your Vercel project settings and add:
- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`
- `CLERK_SECRET_KEY`
- `NEXT_PUBLIC_API_URL` (your Cloud Run backend URL)

3. Deploy:

```bash
vercel
```

Or connect your GitHub repository to Vercel for automatic deployments.

## Environment Configuration

### Backend Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `GCP_PROJECT_ID` | Google Cloud Project ID | Yes | - |
| `PORT` | Server port | No | 8080 |
| `FIRESTORE_EMULATOR_HOST` | Firestore emulator address (dev only) | No | - |

### Frontend Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` | Clerk publishable key | Yes |
| `CLERK_SECRET_KEY` | Clerk secret key | Yes |
| `NEXT_PUBLIC_API_URL` | Backend API URL | Yes |

## Security Considerations

### Production Checklist

- [ ] Use HTTPS for all endpoints
- [ ] Configure CORS appropriately in api.go
- [ ] Set up Cloud Run authentication/authorization
- [ ] Rotate API keys regularly
- [ ] Use Firestore security rules
- [ ] Enable Cloud Run logging and monitoring
- [ ] Set up alerting for errors
- [ ] Configure rate limiting
- [ ] Review Clerk security settings

### Firestore Security Rules

Example security rules:

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Players collection - read for authenticated users
    match /players/{playerId} {
      allow read: if request.auth != null;
      allow write: if request.auth != null && request.auth.uid == playerId;
    }
    
    // Courses collection - read for all, write for admins only
    match /courses/{courseId} {
      allow read: if request.auth != null;
      allow write: if request.auth != null && 
        get(/databases/$(database)/documents/users/$(request.auth.uid)).data.role == 'admin';
    }
    
    // Matches and scores - read for participants
    match /matches/{matchId} {
      allow read: if request.auth != null;
      allow write: if request.auth != null && 
        get(/databases/$(database)/documents/users/$(request.auth.uid)).data.role == 'admin';
    }
    
    match /scores/{scoreId} {
      allow read: if request.auth != null;
      allow write: if request.auth != null;
    }
  }
}
```

## Monitoring and Logging

### Cloud Run

View logs:
```bash
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=golf-league-backend" --limit 50
```

### Set Up Alerts

1. Go to Cloud Console > Monitoring > Alerting
2. Create alert for error rates
3. Set up notification channels (email, Slack, etc.)

## Backup and Recovery

### Firestore Backup

Set up automated backups:

```bash
gcloud firestore backups schedules create \
  --database='(default)' \
  --recurrence=daily \
  --retention=14d
```

## Scaling Considerations

### Cloud Run

- Adjust min/max instances based on load:
  ```bash
  gcloud run services update golf-league-backend \
    --min-instances=1 \
    --max-instances=10
  ```

### Firestore

- Use composite indexes for complex queries
- Monitor read/write operations
- Consider collection group queries for performance

## Cost Optimization

### Cloud Run
- Set min instances to 0 for development
- Use request-based pricing for low traffic
- Monitor and adjust resources based on usage

### Firestore
- Optimize queries to minimize reads
- Use batch operations where possible
- Archive old data periodically

### Vercel
- Free tier supports hobby projects
- Upgrade for team features and higher limits

## Troubleshooting

### Backend Issues

**Container won't start:**
- Check GCP_PROJECT_ID is set correctly
- Verify Firestore is enabled in the project
- Check Cloud Run logs for errors

**API returns 500 errors:**
- Check Firestore permissions
- Verify environment variables
- Review application logs

### Frontend Issues

**Authentication not working:**
- Verify Clerk keys are correct
- Check CORS settings in backend
- Ensure NEXT_PUBLIC_API_URL is accessible

**API calls failing:**
- Verify backend is deployed and accessible
- Check API URL in environment variables
- Review browser network tab for errors

## Support

For deployment issues:
1. Check application logs
2. Review this deployment guide
3. Open an issue on GitHub with details
