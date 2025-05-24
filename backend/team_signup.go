package main

import (
	"database/sql"
	"encoding/json"

	// "log"
	"net/http"
)

type TeamSignupRequest struct {
	CompetitionID int  `json:"competition_id"`
	TeamID        *int `json:"team_id,omitempty"`
}

func NewTeamSignupHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "HTTP Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TeamSignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Check if team exists
		var err error
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE id_team=$1)", *req.TeamID).Scan(&exists)
		if err != nil {
			// log.Printf("Error checking team existence: %v", err)
			http.Error(w, "Team does not exist", http.StatusInternalServerError)
			return
		}
		if req.TeamID == nil {
			http.Error(w, "Team ID is required", http.StatusBadRequest)
			return
		}
		if !exists {
			http.Error(w, "Team does not exist", http.StatusBadRequest)
			return
		}

		// Check if competition exists
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM competitions WHERE competition_id=$1)", req.CompetitionID).Scan(&exists)
		if err != nil {
			// log.Printf("Error checking stage existence: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Competition does not exist", http.StatusBadRequest)
			return
		}

		// Insert into stage_participants
		_, err = db.Exec(`
			INSERT INTO competition_participants (competition_id, team_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, req.CompetitionID, *req.TeamID)

		if err != nil {
			// log.Printf("Error signing up: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Signup successful",
		})
	})
}
