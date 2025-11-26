# Backend TODO: Profile Feature API Requirements

This document outlines the backend API endpoints and data requirements needed to fully support the new Profile feature in the frontend.

## Overview

The frontend Profile page (`/profile/:playerId`) displays:
- Player profile information (name, email, status)
- Leagues the player is enrolled in
- Current and historical handicaps (including provisional at season start)
- Historical rounds within league/season context
- Prior matchups with scorecards and match results

## Current API Status

### Existing Endpoints That Work
- `GET /api/user/me` - Returns current user's player info and linked status ✅
- `GET /api/leagues/{league_id}/seasons` - Lists seasons for a league ✅
- `GET /api/leagues/{league_id}/matches` - Lists all matches ✅
- `GET /api/leagues/{league_id}/members` - Lists league members with player info ✅
- `GET /api/leagues/{league_id}/courses` - Lists courses ✅

### Endpoints Needing Implementation/Verification

#### 1. Player Scores Endpoint
**Endpoint:** `GET /api/leagues/{league_id}/players/{player_id}/scores`

**Current Status:** May exist (`handleGetPlayerScores` in server.go) - needs verification

**Required Response Fields:**
```json
{
  "id": "string",
  "matchId": "string",
  "playerId": "string",
  "leagueId": "string",
  "date": "ISO 8601 string",
  "courseId": "string",
  "holeScores": [int, int, ...], // 9 integers
  "holeAdjustedGrossScores": [int, int, ...], // 9 integers
  "matchNetHoleScores": [int, int, ...], // 9 integers (per-hole net scores)
  "grossScore": int,
  "netScore": int,
  "matchNetScore": int, // Total match net score
  "adjustedGross": int,
  "handicapDifferential": float,
  "handicapIndex": float, // Player's handicap index used/after this round
  "courseHandicap": int,
  "playingHandicap": int,
  "strokesReceived": int,
  "matchStrokes": [int, int, ...], // Strokes per hole
  "playerAbsent": bool
}
```

**Notes:**
- The `date` field is critical for sorting and displaying history
- The `handicapIndex` field should show the handicap after this round was calculated
- The `matchNetHoleScores` are important for displaying detailed scorecards

#### 2. Handicap History Endpoint (Optional Enhancement)
**Endpoint:** `GET /api/leagues/{league_id}/players/{player_id}/handicap-history`

**Currently:** The frontend calculates handicap history from scores. Consider adding a dedicated endpoint for efficiency.

**Suggested Response:**
```json
[
  {
    "date": "ISO 8601 string",
    "handicapIndex": float,
    "isProvisional": bool,
    "seasonId": "string",
    "seasonName": "string",
    "roundId": "string" // null if provisional
  }
]
```

#### 3. Provisional Handicap Data

**Current:** Retrieved from `LeagueMember.provisionalHandicap` field via league members endpoint

**Ensure:** The `provisionalHandicap` field is properly populated when adding players to leagues (per Golf League Rules 3.2)

## Data Model Verification

Ensure the `Score` model in the backend includes these fields that match the frontend type:

```go
type Score struct {
    ID                      string    `firestore:"id" json:"id"`
    MatchID                 string    `firestore:"match_id" json:"matchId"`
    PlayerID                string    `firestore:"player_id" json:"playerId"`
    LeagueID                string    `firestore:"league_id" json:"leagueId"`
    Date                    time.Time `firestore:"date" json:"date"` // CRITICAL
    CourseID                string    `firestore:"course_id" json:"courseId"`
    HoleScores              []int     `firestore:"hole_scores" json:"holeScores"`
    HoleAdjustedGrossScores []int     `firestore:"hole_adjusted_gross_scores" json:"holeAdjustedGrossScores"`
    MatchNetHoleScores      []int     `firestore:"match_net_hole_scores" json:"matchNetHoleScores"`
    GrossScore              int       `firestore:"gross_score" json:"grossScore"`
    NetScore                int       `firestore:"net_score" json:"netScore"`
    MatchNetScore           int       `firestore:"match_net_score" json:"matchNetScore"`
    AdjustedGross           int       `firestore:"adjusted_gross" json:"adjustedGross"`
    HandicapDifferential    float64   `firestore:"handicap_differential" json:"handicapDifferential"`
    HandicapIndex           float64   `firestore:"handicap_index" json:"handicapIndex"` // Index after round
    CourseHandicap          int       `firestore:"course_handicap" json:"courseHandicap"`
    PlayingHandicap         int       `firestore:"playing_handicap" json:"playingHandicap"`
    StrokesReceived         int       `firestore:"strokes_received" json:"strokesReceived"`
    MatchStrokes            []int     `firestore:"match_strokes" json:"matchStrokes"`
    PlayerAbsent            bool      `firestore:"player_absent" json:"playerAbsent"`
}
```

## Security Requirements

The frontend already implements client-side security:
- Only allows viewing own profile (redirects/shows error for other player IDs)

**Backend should also verify:**
1. User can only access their own profile data
2. Scores endpoint should verify the requesting user matches the player_id OR is a league admin

Example validation in `handleGetPlayerScores`:
```go
func (s *APIServer) handleGetPlayerScores(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get authenticated user
    userID, err := GetUserIDFromContext(ctx)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Get player from clerk user ID
    requestingPlayer, err := s.firestoreClient.GetPlayerByClerkID(ctx, userID)
    if err != nil {
        http.Error(w, "Player not found", http.StatusNotFound)
        return
    }
    
    playerID := r.PathValue("id")
    
    // Verify user is requesting their own scores
    if requestingPlayer.ID != playerID {
        // Check if user is league admin (optional: allow admins to view)
        http.Error(w, "Access denied: can only view own scores", http.StatusForbidden)
        return
    }
    
    // Continue with fetching scores...
}
```

## Testing Checklist

- [ ] Verify `GET /api/leagues/{league_id}/players/{player_id}/scores` returns all required fields
- [ ] Verify scores include `date` field properly formatted
- [ ] Verify scores include `handicapIndex` showing handicap after round
- [ ] Verify scores include `matchNetHoleScores` for detailed scorecard display
- [ ] Verify `LeagueMember.provisionalHandicap` is properly set when adding members
- [ ] Test security: non-admin user cannot access another player's scores
- [ ] Test with multiple seasons to ensure season filtering works

## Frontend Integration Notes

The frontend Profile component (`/frontend/src/pages/Profile.tsx`) expects:
1. Scores sorted by date (most recent first for display, chronological for handicap history)
2. Match data to link scores to opponents and courses
3. Season data to filter by season and show season names
4. League member data to display opponent names and roles

All data fetching is done in parallel using `Promise.all` for efficiency.
