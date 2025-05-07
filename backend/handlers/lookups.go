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
				http.Error(w, "Failed to parse sports", http.StatusInternalServerError)
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
				http.Error(w, "Failed to parse structure types", http.StatusInternalServerError)
				return
			}
			list = append(list, i)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}

func GetActivityTypesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT activity_type_id, activity_name FROM activity_types")
		if err != nil {
			http.Error(w, "Failed to fetch activity types", http.StatusInternalServerError)
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
				http.Error(w, "Failed to parse activity types", http.StatusInternalServerError)
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
		rows, err := db.Query("SELECT tourney_format_id, tourney_name FROM tournament_formats")
		if err != nil {
			http.Error(w, "Failed to fetch tournament formats", http.StatusInternalServerError)
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
				http.Error(w, "Failed to parse tournament formats", http.StatusInternalServerError)
				return
			}
			list = append(list, i)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}
