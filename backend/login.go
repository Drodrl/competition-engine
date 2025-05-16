package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

func NewLoginHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "HTTP Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Error reading data", http.StatusBadRequest)
			return
		}

		var userID int
		var roleID int
		err := db.QueryRow(`
			SELECT id_user, role_id 
			FROM users 
			WHERE email = $1 AND password_hash = $2
			`, creds.Email, creds.Password).Scan(&userID, &roleID)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				log.Printf("Database error: %v", err)
			}
			return
		}

		if roleID == 3 {
			response := struct {
				UserID int    `json:"userId"`
				Role   string `json:"role"`
			}{
				UserID: userID,
				Role:   "organizer",
			}

			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, "Error writing response", http.StatusInternalServerError)
				log.Printf("Error encoding response: %v", err)
			}
		} else if roleID == 1 {
			response := struct {
				UserID int    `json:"userId"`
				Role   string `json:"role"`
			}{
				UserID: userID,
				Role:   "admin",
			}

			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, "Error writing response", http.StatusInternalServerError)
				log.Printf("Error encoding response: %v", err)
			}
		} else {
			http.Error(w, "Unauthorized role", http.StatusForbidden)
		}

	})
}
