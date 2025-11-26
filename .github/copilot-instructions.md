# Copilot Instructions for Golf League Manager

This repository contains a full-stack Golf League Manager application with a Go backend API and React/TypeScript frontend.

## Project Overview

The Golf League Manager handles:
- Golf league handicap calculations using USGA-compliant formulas
- 9-hole match play scoring with point distribution (22 points per match)
- Stroke allocation based on playing handicap differences
- Season management and standings tracking

## Technology Stack

### Backend (Go 1.22+)
- **Location**: `/server/`
- **Framework**: Standard library `net/http` with Go 1.22+ routing patterns
- **Database**: Google Cloud Firestore
- **Authentication**: Clerk JWT tokens
- **Testing**: Standard `testing` package with table-driven tests

### Frontend (React/TypeScript)
- **Location**: `/frontend/`
- **Build Tool**: Vite
- **Framework**: React 18 with react-router-dom v6
- **Styling**: Tailwind CSS
- **Authentication**: Clerk React SDK

## Code Conventions

### Go Backend
- Use Go 1.22+ HTTP routing patterns: `s.mux.Handle("GET /api/endpoint", handler)`
- Models use Firestore field tags: `firestore:"field_name" json:"field_name"`
- All business logic belongs in `/server/internal/services/`
- HTTP handlers go in `/server/internal/api/` or `/server/internal/handlers/`
- Use `context.Context` for request-scoped data and timeouts
- Error handling: return errors up the stack, use `fmt.Errorf` for wrapping
- Use structured logging with `log/slog` (Go 1.21+)

### TypeScript Frontend
- Use TypeScript strict mode
- Components use functional style with hooks
- API client in `/frontend/src/lib/api.ts` handles all backend communication
- Types defined in `/frontend/src/types/`
- Use Clerk hooks for authentication (`useAuth`, `useUser`)

### Testing
- Backend: Table-driven tests colocated with implementation (`*_test.go`)
- Run backend tests: `cd server && go test ./...`
- Run frontend lint: `cd frontend && npm run lint`
- Run frontend build: `cd frontend && npm run build`

## Domain Knowledge - Golf Handicap Calculations

### Key Formulas
```
Score Differential = ((Adjusted Gross - Course Rating) × 113) / Slope Rating
Course Handicap = (League Handicap × Slope Rating / 113) + (Course Rating - Par)
Playing Handicap = round(Course Handicap × 0.95)
```

### Established Players (5+ rounds)
- Use last 5 recorded 9-hole rounds
- Drop 2 highest (worst) differentials
- Average remaining 3 lowest differentials
- Round to nearest 0.1

### New Players (< 5 rounds)
- Use committee-assigned provisional handicap
- Apply weighted average with actual rounds
- Net Double Bogey rule applies: `Par + 2 + strokes received on hole`

### Match Play Scoring (22 points per match)
- 2 points per hole (winner takes both, ties split 1-1)
- 4 points for overall lower net total
- Strokes allocated by hole handicap order (1 = hardest)

## Important Implementation Notes

### Handicap Services (`/server/internal/services/`)
- `handicap.go`: Core handicap calculation functions
- `match.go`: Match play scoring and stroke allocation
- `jobs.go`: Background jobs for recalculation

### Data Models (`/server/internal/models/`)
- Player, Course, Season, Match, Score, HandicapRecord
- Completed matches cannot be edited (status check in handlers)

### API Structure
- Public: `/health`, `/health/ready`
- Authenticated: `/api/user/*`
- League-scoped: `/api/leagues/{league_id}/*`
- All authenticated endpoints require Clerk JWT Bearer token

## Build and Test Commands

```bash
# Backend
cd server
go test ./...              # Run all tests
go test -v ./...           # Verbose output
go vet ./...               # Static analysis
gofmt -l .                 # Check formatting

# Frontend
cd frontend
npm install                # Install dependencies
npm run build              # Build for production
npm run lint               # Run ESLint
npm run dev                # Development server
```

## Environment Variables

### Backend
- `GCP_PROJECT_ID` - Google Cloud project for Firestore
- `CLERK_SECRET_KEY` - Clerk API secret
- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - dev, staging, production
- `CORS_ORIGINS` - Allowed origins (comma-separated)

### Frontend
- `VITE_CLERK_PUBLISHABLE_KEY` - Clerk publishable key
- `VITE_API_URL` - Backend API URL
