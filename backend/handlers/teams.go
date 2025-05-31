package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Drodrl/competition-engine/models"
)

func NewTeamsHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			log.Println("User ID is required")
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Get teams where the user participates
		rows, err := db.Query(`
            SELECT t.team_id, t.team_name
            FROM user_teams ut
            JOIN teams t ON ut.team_id = t.team_id
            WHERE ut.user_id = $1
        `, userID)

		if err != nil {
			log.Printf("Error fetching teams for user %s: %v", userID, err)
			http.Error(w, "Failed to fetch teams", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var teams []models.Team
		for rows.Next() {
			var team models.Team
			if err := rows.Scan(&team.ID, &team.TeamName); err != nil {
				log.Printf("Error scanning team for user %s: %v", userID, err)
				http.Error(w, "Failed to scan team", http.StatusInternalServerError)
				return
			}
			teams = append(teams, team)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(teams); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
	})
}

func RemoveParticipantsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			TeamID  int   `json:"team_id"`
			UserIDs []int `json:"user_ids"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				log.Printf("Error during rollback: %v", err)
			}
		}()

		// Remove users from the team
		for _, userID := range payload.UserIDs {
			_, err := tx.Exec("DELETE FROM user_teams WHERE team_id = $1 AND user_id = $2", payload.TeamID, userID)
			if err != nil {
				http.Error(w, "Failed to remove participants", http.StatusInternalServerError)
				return
			}
		}

		// Update the date_updated field in the teams table
		_, err = tx.Exec("UPDATE teams SET date_updated = NOW() WHERE team_id = $1", payload.TeamID)
		if err != nil {
			http.Error(w, "Failed to update team date", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func AddParticipantsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			TeamID  int   `json:"team_id"`
			UserIDs []int `json:"user_ids"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("Error decoding payload: %v", err)
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		log.Printf("Payload received: %+v", payload)

		tx, err := db.Begin()
		if err != nil {
			log.Printf("Error starting transaction: %v", err)
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				log.Printf("Error during rollback: %v", err)
			}
		}()

		// Add users to the team
		for _, userID := range payload.UserIDs {
			// Check if the user exists in the users table
			var exists bool
			err := tx.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE id_user = $1)", userID).Scan(&exists)
			if err != nil {
				log.Printf("Error checking if user exists (user_id: %d): %v", userID, err)
				http.Error(w, "Failed to validate user", http.StatusInternalServerError)
				return
			}

			if !exists {
				log.Printf("User does not exist (user_id: %d)", userID)
				// http.Error(w, "User does not exist", http.StatusBadRequest)
				sendJSONError(w, "User does not exist", http.StatusBadRequest)
				return
			}
			// Check if the user is already in the team
			exists = false
			err = tx.QueryRow("SELECT EXISTS (SELECT 1 FROM user_teams WHERE team_id = $1 AND user_id = $2)", payload.TeamID, userID).Scan(&exists)
			if err != nil {
				log.Printf("Error checking if user exists (team_id: %d, user_id: %d): %v", payload.TeamID, userID, err)
				http.Error(w, "Failed to check existing participants", http.StatusInternalServerError)
				return
			}

			if !exists {
				_, err := tx.Exec("INSERT INTO user_teams (user_id, team_id, date_updated) VALUES ($1, $2, NOW())", userID, payload.TeamID)
				if err != nil {
					log.Printf("Error inserting user (team_id: %d, user_id: %d): %v", payload.TeamID, userID, err)
					http.Error(w, "Failed to add participants", http.StatusInternalServerError)
					return
				}
			}
		}

		// Update the date_updated field in the teams table
		_, err = tx.Exec("UPDATE teams SET date_updated = NOW() WHERE team_id = $1", payload.TeamID)
		if err != nil {
			log.Printf("Error updating team date (team_id: %d): %v", payload.TeamID, err)
			http.Error(w, "Failed to update team date", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Printf("Error committing transaction: %v", err)
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
