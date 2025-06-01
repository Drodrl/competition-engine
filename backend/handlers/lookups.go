package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Drodrl/competition-engine/models"
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
		if err := json.NewEncoder(w).Encode(list); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
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
		if err := json.NewEncoder(w).Encode(list); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
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
		if err := json.NewEncoder(w).Encode(list); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func GetUserTeamsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(`
            SELECT t.team_id, t.team_name
            FROM user_teams ut
            INNER JOIN teams t ON ut.team_id = t.team_id
            WHERE ut.user_id = $1
        `, userID)
		if err != nil {
			http.Error(w, "Failed to fetch user teams", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var teams []models.Team
		for rows.Next() {
			var team models.Team
			if err := rows.Scan(&team.ID, &team.TeamName); err != nil {
				http.Error(w, "Failed to scan team", http.StatusInternalServerError)
				return
			}
			teams = append(teams, team)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(teams); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func GetTeamParticipantsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teamID := r.URL.Query().Get("team_id")
		if teamID == "" {
			http.Error(w, "Team ID is required", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(`
            SELECT u.id_user, u.name_user, u.lname1_user
            FROM user_teams ut
            INNER JOIN users u ON ut.user_id = u.id_user
            WHERE ut.team_id = $1
        `, teamID)
		if err != nil {
			http.Error(w, "Failed to fetch team participants", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var participants []models.User
		for rows.Next() {
			var user models.User
			if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName); err != nil {
				http.Error(w, "Failed to scan participant", http.StatusInternalServerError)
				return
			}
			participants = append(participants, user)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(participants); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
