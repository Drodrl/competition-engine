package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Drodrl/competition-engine/models"
)

func NewAthletesHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id_user, name_user, lname1_user FROM users WHERE role_id = 2")
		if err != nil {
			http.Error(w, "Failed to fetch athletes", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var u models.User
			if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName); err != nil {
				http.Error(w, "Failed to scan user", http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(users); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
	})
}
