package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type Competition struct {
	ID        int    `json:"competition_id"`
	Name      string `json:"competition_name"`
	Sport     string `json:"sport_id"`
	StartDate string `json:"start_date"`
}

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

		var competitions []Competition
		for rows.Next() {
			var c Competition
			if err := rows.Scan(&c.ID, &c.Name, &c.Sport, &c.StartDate); err != nil {
				http.Error(w, "Failed to scan competition"+err.Error(), http.StatusInternalServerError)
				return
			}
			competitions = append(competitions, c)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(competitions)
	})
}
