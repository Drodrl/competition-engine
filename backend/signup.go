package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type SignupRequest struct {
	StageID int  `json:"stage_id"`
	UserID  *int `json:"user_id,omitempty"`
	TeamID  *int `json:"team_id,omitempty"`
}

func NewSignupHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "HTTP Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if (req.UserID == nil && req.TeamID == nil) || (req.UserID != nil && req.TeamID != nil) {
			http.Error(w, "Must provide either user_id or team_id", http.StatusBadRequest)
			return
		}

		var err error
		var exists bool

		if req.UserID != nil {
			// Check if user exists
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)", *req.UserID).Scan(&exists)
			if err != nil {
				log.Printf("Error checking user existence: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !exists {
				http.Error(w, "User does not exist", http.StatusBadRequest)
				return
			}
			_, err = db.Exec(`
                INSERT INTO stage_participants (stage_id, user_id)
                VALUES ($1, $2)
                ON CONFLICT DO NOTHING
            `, req.StageID, *req.UserID)
		} else {
			// Check if team exists
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE id=$1)", *req.TeamID).Scan(&exists)
			if err != nil {
				log.Printf("Error checking team existence: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !exists {
				http.Error(w, "Team does not exist", http.StatusBadRequest)
				return
			}

			_, err = db.Exec(`
                INSERT INTO stage_participants (stage_id, team_id)
                VALUES ($1, $2)
                ON CONFLICT DO NOTHING
            `, req.StageID, *req.TeamID)
		}

		if err != nil {
			log.Printf("Error signing up: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Signup successful",
		})
	})
}
