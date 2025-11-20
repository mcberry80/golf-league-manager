# Golf League Manager - Quick Start Guide

## 🎯 What's New

✅ **Premium Modern Styling** - Glassmorphism, gradients, smooth animations
✅ **All Navigation Fixed** - No more 404 errors!
✅ **Complete Admin Dashboard** - Manage everything from one place
✅ **Score Entry with Absence Handling** - Automatic handicap adjustment
✅ **Ready for Production** - Deployment scripts included

---

## 🚀 Quick Start (Local Development)

### 1. Setup Environment Variables

**Frontend** (`frontend/.env.local`):
```bash
VITE_CLERK_PUBLISHABLE_KEY=your_clerk_publishable_key
VITE_API_URL=http://localhost:8080
```

**Backend** (PowerShell):
```powershell
$env:GCP_PROJECT_ID="your-gcp-project-id"
$env:CLERK_SECRET_KEY="your_clerk_secret_key"
```

### 2. Start Backend

```powershell
cd server\cmd
go run main.go
```

Backend runs on: http://localhost:8080

### 3. Start Frontend

```powershell
cd frontend
npm install
npm run dev
```

Frontend runs on: http://localhost:5173

### 4. Access Application

1. Open http://localhost:5173
2. Sign in with Clerk
3. Go to "Link Account" and enter your email
4. In Firestore, set `is_admin: true` on your player document
5. Access Admin Dashboard

---

## 📱 Application Features

### Admin Functions
- **League Setup**: Create and manage seasons
- **Player Management**: Add players, set admin privileges
- **Course Management**: Add courses with hole details
- **Match Scheduling**: Schedule weekly matchups
- **Score Entry**: Enter scores with absence handling
- **View Standings**: See league rankings

### Player Functions
- **My Profile**: View handicap and round history
- **Standings**: View league standings
- **Link Account**: Connect Clerk account to player profile

---

## 🎨 New Design Features

- **Dark Theme**: Modern dark background with vibrant accents
- **Glassmorphism**: Translucent cards with backdrop blur
- **Smooth Animations**: Fade-in, hover effects, transitions
- **Premium Components**: Gradient buttons, styled forms, responsive tables
- **Mobile Responsive**: Works on all screen sizes

---

## 🔧 Admin Workflow

1. **Create Season**: Admin → League Setup → Create "Fall 2024"
2. **Add Players**: Admin → Players → Add Allison, Bob, Charlie, David
3. **Add Courses**: Admin → Courses → Add Pine Valley, Oak Ridge
4. **Schedule Matches**: Admin → Match Scheduling → Create matchups
5. **Enter Scores**: Admin → Score Entry → Select match, enter scores
6. **View Results**: Standings → See updated points and rankings

---

## 🚢 Deploy to GCP

Run the deployment script:

```powershell
.\deploy.ps1 `
    -ProjectId "your-gcp-project-id" `
    -ClerkPublishableKey "pk_test_..." `
    -ClerkSecretKey "sk_test_..." `
    -Region "us-central1"
```

See [DEPLOYMENT.md](file:///c:/Dev/golf-league-manager/DEPLOYMENT.md) for detailed instructions.

---

## 📚 Documentation

- [DEPLOYMENT.md](file:///c:/Dev/golf-league-manager/DEPLOYMENT.md) - Deployment guide
- [README.md](file:///c:/Dev/golf-league-manager/README.md) - Full project documentation
- [Golf League Rules.md](file:///c:/Dev/golf-league-manager/Golf%20League%20Rules.md) - League rules and scoring

---

## ✨ Key Improvements

1. **No More 404s**: All routes configured and working
2. **Premium UI**: Modern, professional design throughout
3. **Absence Handling**: Automatic handicap adjustment per league rules
4. **Type Safety**: Full TypeScript integration
5. **Ready to Deploy**: One-command deployment to GCP

---

## 🐛 Troubleshooting

**404 Errors**: All routes are now configured. Clear browser cache if issues persist.

**CORS Issues**: Backend allows all origins. Check Cloud Run logs if issues occur.

**Auth Issues**: Verify Clerk keys and add frontend URL to Clerk's allowed origins.

**Firestore Errors**: Ensure Firestore is enabled and service account has permissions.

---

## 📞 Next Steps

1. Test locally following the Quick Start guide
2. Create test data using the admin UI
3. Test score entry with absence scenario
4. Deploy to GCP when ready
5. Update Clerk settings with production URLs

Enjoy your new Golf League Manager! 🏌️‍♂️⛳
