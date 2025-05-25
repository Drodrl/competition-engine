package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func GetSportsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT sport_id, sport_name FROM sports")
		if err != nil {
			http.Error(w, "Failed to fetch sports", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		type Item struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		var list []Item
		for rows.Next() {
			var i Item
			if err := rows.Scan(&i.ID, &i.Name); err != nil {
				http.Error(w, "Failed to scan sport", http.StatusInternalServerError)
				return
			}
			list = append(list, i)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}

func GetStructureTypesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT structure_type_id, structure_name FROM structure_types")
		if err != nil {
			http.Error(w, "Failed to fetch structure types", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		type Item struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		var list []Item
		for rows.Next() {
			var i Item
			if err := rows.Scan(&i.ID, &i.Name); err != nil {
				http.Error(w, "Failed to scan structure type", http.StatusInternalServerError)
				return
			}
			list = append(list, i)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}

func GetTournamentFormatsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT tourney_format_id, tourney_name, min_participants FROM tournament_formats")
		if err != nil {
			http.Error(w, "Failed to fetch tournament formats", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		type TournamentFormat struct {
			ID              int    `json:"id"`
			Name            string `json:"name"`
			MinParticipants int    `json:"min_participants"`
		}
		var list []TournamentFormat
		for rows.Next() {
			var i TournamentFormat
			if err := rows.Scan(&i.ID, &i.Name, &i.MinParticipants); err != nil {
				http.Error(w, "Failed to scan tournament format", http.StatusInternalServerError)
				return
			}
			list = append(list, i)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}
