package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

type StandingsEntry struct {
	PlayerID      string `json:"playerId"`
	PlayerName    string `json:"playerName"`
	MatchesPlayed int    `json:"matchesPlayed"`
	MatchesWon    int    `json:"matchesWon"`
	MatchesLost   int    `json:"matchesLost"`
	MatchesTied   int    `json:"matchesTied"`
	TotalPoints   int    `json:"totalPoints"`
}

func (s *APIServer) handleGetStandings(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("league_id")
	if leagueID == "" {
		http.Error(w, "League ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	members, err := s.firestoreClient.ListLeagueMembers(ctx, leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get league members: %v", err), http.StatusInternalServerError)
		return
	}

	matches, err := s.firestoreClient.ListMatches(ctx, leagueID, "completed")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get matches: %v", err), http.StatusInternalServerError)
		return
	}

	standingsMap := make(map[string]*StandingsEntry)

	for _, member := range members {
		player, err := s.firestoreClient.GetPlayer(ctx, member.PlayerID)
		if err != nil {
			continue
		}

		entry := &StandingsEntry{
			PlayerID:   player.ID,
			PlayerName: player.Name,
		}
		standingsMap[player.ID] = entry
	}

	for _, match := range matches {
		if match.PlayerAPoints == 0 && match.PlayerBPoints == 0 {
			continue
		}

		if entryA, ok := standingsMap[match.PlayerAID]; ok {
			entryA.MatchesPlayed++
			entryA.TotalPoints += match.PlayerAPoints
			if match.PlayerAPoints > match.PlayerBPoints {
				entryA.MatchesWon++
			} else if match.PlayerAPoints < match.PlayerBPoints {
				entryA.MatchesLost++
			} else {
				entryA.MatchesTied++
			}
		}

		if entryB, ok := standingsMap[match.PlayerBID]; ok {
			entryB.MatchesPlayed++
			entryB.TotalPoints += match.PlayerBPoints
			if match.PlayerBPoints > match.PlayerAPoints {
				entryB.MatchesWon++
			} else if match.PlayerBPoints < match.PlayerAPoints {
				entryB.MatchesLost++
			} else {
				entryB.MatchesTied++
			}
		}
	}

	standings := make([]StandingsEntry, 0, len(standingsMap))
	for _, entry := range standingsMap {
		standings = append(standings, *entry)
	}

	sort.Slice(standings, func(i, j int) bool {
		return standings[i].TotalPoints > standings[j].TotalPoints
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(standings)
}