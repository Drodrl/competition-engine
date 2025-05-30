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
