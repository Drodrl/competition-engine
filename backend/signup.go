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

		if req.UserID == nil && req.TeamID == nil {
			http.Error(w, "Invalid input: Provide an user_id or team_id", http.StatusBadRequest)
			return
		}

		query := `
            INSERT INTO stage_participants (stage_id, user_id, team_id)
            VALUES ($1, $2, $3)
        `
		_, err := db.Exec(query, req.StageID, req.UserID, req.TeamID)
		if err != nil {
			http.Error(w, "Failed to save competition signup: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Error saving signup: %v", err)
			return
		}

		log.Println("INFO: Signup successful for stage_id:", req.StageID, "user_id:", req.UserID, "team_id:", req.TeamID)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Signup successful"))
	})
}
