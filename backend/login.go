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

		response := LoginResponse{
			Message: "Login successful",
			Token:   "token-dummy", //Token dummy used as a placeholder
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			log.Printf("Error encoding response: %v", err)
		}
	})
}
