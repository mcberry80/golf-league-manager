# Golf League Manager

A comprehensive golf league scoring and handicap system built with Go, Firestore, and designed to run on Google Cloud Run.

## Features

- **Handicap Calculation**: Automatic calculation of league, course, and playing handicaps using USGA-compliant formulas
- **Match Play Scoring**: Full support for 9-hole match play with stroke allocation and point calculation
- **Adjusted Gross Scoring**: Net Double Bogey rule for established players, par + 5 cap for new players
- **Absence Policy**: Automatic handicap adjustment for absent players
- **Rules Engine**: Built-in support for breakfast ball, gimmes, hazards, and lie improvement rules
- **Firestore Integration**: Complete CRUD operations for all entities
- **Automated Jobs**: Weekly handicap recalculation and match completion processing

## Data Models

### Player
- ID, Name, Email
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
- Week number and two player references
- Course reference, date, and status

### Score
- Match and player references
- Hole number, gross/net scores, strokes received

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
- Go 1.23.2 or later
- Google Cloud project with Firestore enabled
- GCP credentials configured

### Installation

```bash
# Clone the repository
git clone https://github.com/mcberry80/golf-league-manager.git
cd golf-league-manager

# Install dependencies
go mod download

# Run tests
go test ./...
```

### Environment Variables

```bash
# GCP Project ID for Firestore
export GCP_PROJECT_ID="your-project-id"

# For local development with Firestore emulator
export FIRESTORE_EMULATOR_HOST="localhost:8080"
```

### Running the Handicap Recalculation Job

```go
import (
    "context"
    glm "golf-league-manager"
)

func main() {
    ctx := context.Background()
    
    // Create Firestore client
    fc, err := glm.NewFirestoreClient(ctx, "your-project-id")
    if err != nil {
        log.Fatal(err)
    }
    defer fc.Close()
    
    // Run handicap recalculation
    job := glm.NewHandicapRecalculationJob(fc)
    if err := job.Run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

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

## API Usage Examples

### Create a Player

```go
player := glm.Player{
    ID:          uuid.New().String(),
    Name:        "John Doe",
    Email:       "john@example.com",
    Active:      true,
    Established: false,
    CreatedAt:   time.Now(),
}

err := fc.CreatePlayer(ctx, player)
```

### Record a Round

```go
round := glm.Round{
    ID:          uuid.New().String(),
    PlayerID:    "player-id",
    Date:        time.Now(),
    CourseID:    "course-id",
    GrossScores: []int{5, 4, 6, 5, 4, 3, 7, 5, 4},
}

// Process the round to calculate adjusted scores
processor := glm.NewMatchCompletionProcessor(fc)
err := processor.ProcessRound(ctx, round.ID)
```

### Calculate Match Points

```go
scoresA, _ := fc.GetPlayerMatchScores(ctx, matchID, playerAID)
scoresB, _ := fc.GetPlayerMatchScores(ctx, matchID, playerBID)
course, _ := fc.GetCourse(ctx, courseID)

pointsA, pointsB := glm.CalculateMatchPoints(scoresA, scoresB, *course)
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

### Modules

- **models.go** - Data structures for all entities
- **handicap.go** - Handicap calculation logic
- **match.go** - Match play scoring and stroke allocation
- **rules.go** - Gameplay rule validators and helpers
- **persistence.go** - Firestore CRUD operations
- **jobs.go** - Background jobs for recalculation and processing

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
