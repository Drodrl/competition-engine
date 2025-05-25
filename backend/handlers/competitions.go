package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Drodrl/competition-engine/models"
)

func NewCompetitionListHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
            SELECT c.competition_id, c.competition_name, s.sport_name, c.start_date
            FROM competitions c
            JOIN sports s ON c.sport_id = s.sport_id
        `)
		if err != nil {
			http.Error(w, "Failed to fetch competitions", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var competitions []models.Competition
		for rows.Next() {
			var c models.Competition
			if err := rows.Scan(&c.ID, &c.Name, &c.Sport, &c.StartDate); err != nil {
				http.Error(w, "Failed to scan competition"+err.Error(), http.StatusInternalServerError)
				return
			}
			competitions = append(competitions, c)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(competitions); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
	})

}
