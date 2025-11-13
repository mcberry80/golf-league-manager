# Golf League Manager

A comprehensive golf league scoring and handicap system with Go backend API, React/TypeScript frontend, and Firestore database, designed to run on Google Cloud Run.

## Features

### Backend (Go)
- **Clerk Authentication**: Secure API authentication using Clerk JWT tokens
- **Role-Based Authorization**: Admin-only endpoints and league member access controls
- **Season Management**: Full season scheduling with ability to setup and manage league seasons, including match schedules throughout the season
- **Match Protection**: Prevents editing of completed matches while allowing full control over future/scheduled matches
- **Immediate Handicap Updates**: Handicaps are automatically recalculated immediately after each round is entered (no longer requires scheduled job)
- **Handicap Calculation**: Automatic calculation of league, course, and playing handicaps using USGA-compliant formulas
- **Match Play Scoring**: Full support for 9-hole match play with stroke allocation and point calculation
- **Adjusted Gross Scoring**: Net Double Bogey rule for established players, par + 5 cap for new players
- **Absence Policy**: Automatic handicap adjustment for absent players
- **REST API**: Complete web services for admin operations and player queries using Go 1.22+ routing
- **Firestore Integration**: Complete CRUD operations for all entities
- **Automated Jobs**: Weekly handicap recalculation and match completion processing

### Frontend (React/TypeScript)
- **Clerk Authentication**: Secure user authentication with email/password and social login
- **Account Linking**: Players can link their Clerk account to their league profile using their email
- **Admin Dashboard**: Manage courses, players, matches, and enter scores (admin-only)
- **Player Portal**: View personal scores, handicaps, and match history (league members)
- **League Standings**: Real-time standings and player rankings (league members)
- **Responsive Design**: Modern UI with Tailwind CSS

## Data Models

### Season
- ID, Name
- Start date and end date
- Active status (only one season can be active at a time)
- Description and creation timestamp

### Player
- ID, Name, Email
- Clerk User ID (for authentication linking)
- Admin status (for role-based access control)
- Active status and Established status (5+ rounds)
- Creation timestamp

### Round
- Player reference and Course reference
- 9-hole gross scores and adjusted scores
- Date and totals

### Course
- Name, Par, Course Rating, Slope Rating
- Hole pars and handicaps (difficulty rankings 1-9)

### Handicap Record
- Player reference
- League, Course, and Playing handicaps
- Update timestamp (for audit trail)

### Match
- Season reference (linking match to a specific season)
- Week number and two player references
- Course reference, date, and status (scheduled|completed)
- **Note**: Completed matches cannot be edited

### Score
- Match and player references
- Hole number, gross/net scores, strokes received

## Project Structure

The project follows a clean, modular architecture with all business logic organized in the server structure:

```
golf-league-manager/
├── main.go                    # Root package declaration
├── server/                    # Server application
│   ├── cmd/                   # Server entry point
│   │   └── main.go           # Server startup
│   └── internal/             # Internal server packages
│       ├── api/              # HTTP server and routing
│       │   └── server.go     # API server implementation
│       ├── handlers/         # HTTP request handlers (future)
│       ├── models/           # Data models and types
│       │   └── models.go     # Player, Round, Course, etc.
│       ├── persistence/      # Database layer
│       │   └── firestore.go  # Firestore operations
│       └── services/         # Business logic services
│           ├── handicap.go   # Core handicap calculations
│           ├── match.go      # Match processing
│           ├── jobs.go       # Background jobs
│           └── *_test.go     # Service tests
└── frontend/                  # React/TypeScript frontend
    └── src/
        ├── app/              # Next.js app pages
        ├── lib/              # Frontend utilities
        └── types/            # TypeScript types
```

**Architecture Principles:**
- **Business Logic in Services**: All handicap calculations and business logic in `server/internal/services`
- **Modular Organization**: Server code organized into clear, focused packages
- **Consistent Structure**: Server follows similar organizational pattern as frontend
- **Separation of Concerns**: Models, persistence, services, and API layers are separated
- **Testability**: Each layer has corresponding test files co-located with implementation

## Handicap Rules

### League Handicap Calculation
1. Use player's last 5 rounds
2. Calculate score differential for each: `((adjusted_gross - course_rating) * 113) / slope_rating`
3. Drop the two highest differentials
4. Average the remaining three
5. Round to 0.1

### New Players (< 5 rounds)
- Use average of available differentials
- First 3 matches: add +2 strokes (provisional adjustment)
- Scores capped at par + 5 per hole

### Established Players (5+ rounds)
- Apply Net Double Bogey rule: `min(gross, par + 2 + strokes_received)`
- Full handicap calculation with drop rules

### Course & Playing Handicap
```go
course_handicap = (league_handicap * slope_rating / 113) + (course_rating - par)
playing_handicap = round(course_handicap * 0.95)
```

### Absence Policy
```go
absent_handicap = max(posted_handicap + 2, average_of_worst_3_from_last_5)
// Capped at posted_handicap + 4
```

## Match Play Rules

### Point Distribution (22 points total)
- 2 points per hole (winner gets 2, tie = 1-1 split)
- 4 points for overall lower net total

### Stroke Allocation
- Only the higher-handicap player receives strokes
- Strokes allocated by hole handicap order (1 = hardest → 9 = easiest)
- Difference in playing handicaps determines stroke count

### Gameplay Rules

**Breakfast Ball**
- Allowed only on hole 1
- If used, must use the 2nd shot

**Out of Bounds / Lost Ball**
- +1 stroke penalty
- Drop near loss point (not closer to hole) or retee as "hitting 3"

**Hazards (Penalty Areas)**
- Crossing hazards: Drop behind entry point on line with flag (+1 stroke)
- Lateral hazards: Drop within 2 club lengths of entry, no closer to hole (+1 stroke)
- Use true entry point (not line of flight)

**Fluff Rule (Lie Improvement)**
- Ball may be moved within 3 inches using clubhead
- Cannot eliminate obstacles (rocks, roots, etc.)

**Gimmes**
- Only for putts ≤ 2 feet
- All other putts must be holed out

## Setup

### Prerequisites
- Go 1.22 or later
- Node.js 18+ (for frontend)
- Google Cloud project with Firestore enabled
- Clerk account for authentication (get one at [clerk.com](https://clerk.com))
- GCP credentials configured

### Backend Setup

```bash
# Clone the repository
git clone https://github.com/mcberry80/golf-league-manager.git
cd golf-league-manager

# Install Go dependencies
go mod download

# Run tests
go test ./...

# Start the API server
cd server/cmd
go run main.go
```

### Frontend Setup

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Copy environment template
cp .env.local.example .env.local

# Edit .env.local with your Clerk keys and API URL

# Start development server
npm run dev
```

### Environment Variables

**Backend (.env or export)**
```bash
# GCP Project ID for Firestore
export GCP_PROJECT_ID="your-project-id"

# Clerk Secret Key (required for JWT verification)
export CLERK_SECRET_KEY="sk_test_..."

# Server port (default: 8080)
export PORT="8080"

# For local development with Firestore emulator
export FIRESTORE_EMULATOR_HOST="localhost:8080"
```

**Frontend (frontend/.env.local)**
```bash
# Clerk Authentication
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...

# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Running the Backend Server

```bash
# Set environment variables
export GCP_PROJECT_ID="your-project-id"
export PORT="8080"

# Run the server
cd server/cmd
go run main.go
```

The API server will be available at `http://localhost:8080`

### Running the Frontend

```bash
cd frontend
npm run dev
```

The web application will be available at `http://localhost:3000`

### Cloud Run Deployment

The system is designed to run on Google Cloud Run. Example deployment:

```bash
# Build container
gcloud builds submit --tag gcr.io/PROJECT-ID/golf-league-manager

# Deploy to Cloud Run
gcloud run deploy golf-league-manager \
  --image gcr.io/PROJECT-ID/golf-league-manager \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

### Cloud Scheduler Setup

For weekly handicap recalculation:

```bash
# Create a Cloud Scheduler job to trigger weekly recalculation
gcloud scheduler jobs create http handicap-recalc \
  --schedule="0 2 * * 1" \
  --uri="https://your-cloud-run-url/api/jobs/recalculate-handicaps" \
  --http-method=POST \
  --time-zone="America/New_York"
```

## Authentication & Authorization

### Overview

The Golf League Manager uses Clerk for authentication and implements role-based access control:

- **Admin Role**: Full access to all admin endpoints (creating/editing courses, players, matches, scores)
- **League Member**: Access to view league data (standings, matches, player stats)
- **Authentication Required**: All API endpoints (except health check) require a valid Clerk JWT token

### Account Linking Flow

1. **Admin Creates Player**: Admin adds a player to the league with their email address
2. **User Signs Up**: User creates a Clerk account (can be email/password, Google, etc.)
3. **Link Account**: User visits `/link-account` and enters the email used by the admin
4. **Verification**: System links the Clerk user ID to the player profile
5. **Access Granted**: User can now access league features based on their role

### API Authentication

All authenticated endpoints require an `Authorization` header with a Clerk JWT token:

```bash
Authorization: Bearer <clerk-jwt-token>
```

The frontend automatically includes this header when using the authenticated API client.

### Setting Admin Role

To grant admin access to a player:

1. Update the player document in Firestore
2. Set `is_admin: true` for the player
3. The user will have admin access on their next API request

## API Endpoints

The backend exposes RESTful endpoints using Go 1.22+ routing with method and path matching:

**Note**: All endpoints except `/health` require authentication via Clerk JWT token in the `Authorization` header.

### User Account Endpoints

**Authentication Required**

- `GET /api/user/me` - Get current user information and player link status
- `POST /api/user/link-player` - Link Clerk account to player profile by email

### Admin Endpoints

**Requires Authentication + Admin Role**

**Courses**
- `POST /api/admin/courses` - Create a course
- `GET /api/admin/courses` - List all courses
- `GET /api/admin/courses/{id}` - Get course by ID
- `PUT /api/admin/courses/{id}` - Update course

**Players**
- `POST /api/admin/players` - Create a player
- `GET /api/admin/players?active=true` - List players (optionally filter by active)
- `GET /api/admin/players/{id}` - Get player by ID
- `PUT /api/admin/players/{id}` - Update player

**Seasons**
- `POST /api/admin/seasons` - Create a season
- `GET /api/admin/seasons` - List all seasons (ordered by start date)
- `GET /api/admin/seasons/{id}` - Get season by ID
- `PUT /api/admin/seasons/{id}` - Update season
- `GET /api/admin/seasons/active` - Get the currently active season
- `GET /api/admin/seasons/{id}/matches` - Get all matches for a season

**Matches**
- `POST /api/admin/matches` - Create a match
- `GET /api/admin/matches?status=scheduled` - List matches (optionally filter by status)
- `GET /api/admin/matches/{id}` - Get match by ID
- `PUT /api/admin/matches/{id}` - Update match (only for scheduled matches, not completed)

**Scores & Rounds**
- `POST /api/admin/scores` - Enter a score for a hole
- `POST /api/admin/rounds` - Create a round (automatically processes adjusted scores and recalculates handicap)

**Jobs**
- `POST /api/jobs/recalculate-handicaps` - Trigger handicap recalculation for all players
- `POST /api/jobs/process-match/{id}` - Process a completed match

### League Member Endpoints

**Requires Authentication + League Membership**

- `GET /api/players/{id}/handicap` - Get player's current handicap
- `GET /api/players/{id}/rounds` - Get player's round history
- `GET /api/matches/{id}/scores` - Get all scores for a match
- `GET /api/standings` - Get league standings

### Frontend API Usage

**Using the authenticated API client:**

```typescript
'use client'

import { useAuthenticatedAPI } from '@/lib/useAuthenticatedAPI'

function MyComponent() {
  const { api, isReady } = useAuthenticatedAPI()

  useEffect(() => {
    if (isReady) {
      // API client is configured with Clerk token
      loadData()
    }
  }, [isReady])

  const loadData = async () => {
    // Automatically includes Authorization header
    const standings = await api.getStandings()
    
    // Admin endpoints require admin role
    const courses = await api.listCourses()
  }
}
```

**API Examples:**

```typescript
import { api } from '@/lib/api'

// Link user account to player profile
await api.linkPlayerAccount('user@example.com')

// Get current user information
const userInfo = await api.getCurrentUser()

// Create a course (admin only)
const course = await api.createCourse({
  name: 'Pine Valley',
  par: 36,
  course_rating: 35.5,
  slope_rating: 113,
  hole_handicaps: [1, 2, 3, 4, 5, 6, 7, 8, 9],
  hole_pars: [4, 3, 5, 4, 4, 3, 5, 4, 4]
})

// Get standings (league members)
const standings = await api.getStandings()

// Enter a score
await api.enterScore({
  match_id: 'match-123',
  player_id: 'player-456',
  hole_number: 1,
  gross_score: 5,
  net_score: 4,
  strokes_received: 1
})
```

## Testing

All core functionality includes comprehensive unit tests:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Test Structure
- `main_test.go` - Original handicap calculation tests
- `handicap_test.go` - Handicap calculation and adjustment tests
- `match_test.go` - Match play scoring and stroke allocation tests
- `rules_test.go` - Gameplay rules validation tests

## Architecture

### Backend Modules

- **models.go** - Data structures for all entities with Firestore tags
- **handicap.go** - Handicap calculation logic
- **match.go** - Match play scoring and stroke allocation
- **persistence.go** - Firestore CRUD operations
- **jobs.go** - Background jobs for recalculation and processing
- **api.go** - REST API endpoints using Go 1.22+ routing
- **cmd/server/main.go** - HTTP server entry point

### Frontend Structure

- **src/app/** - Next.js app router pages
  - **page.tsx** - Home page with authentication
  - **admin/** - Admin dashboard and management pages
  - **standings/** - League standings page
  - **players/** - Player profile pages
- **src/lib/api.ts** - API client for backend communication
- **src/types/** - TypeScript type definitions
- **src/components/** - Reusable React components

### Design Principles

1. **Idiomatic Go** - Clean, simple, and maintainable code
2. **Modular Design** - Each module has a single responsibility
3. **Firestore Best Practices** - Efficient queries and proper indexing
4. **Comprehensive Testing** - Table-driven tests for all logic
5. **Audit Trail** - All handicap changes tracked with timestamps

## Transparency

The system provides full transparency by:
- Storing all handicap calculation history
- Recording last 5 differentials for each player
- Identifying which 3 differentials were used
- Tracking when handicaps were updated
- Maintaining audit logs for all calculations

## License

Copyright (c) 2024 mcberry80

## Support

For issues, questions, or contributions, please open an issue on GitHub.
