package handlers

import (
	"database/sql"
	"encoding/json"

	// "log"
	"net/http"
)

type UserSignupRequest struct {
	CompetitionID int  `json:"competition_id"`
	UserID        *int `json:"user_id,omitempty"`
}

func NewUserSignupHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "HTTP Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req UserSignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Check if user exists
		var err error
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id_user=$1)", *req.UserID).Scan(&exists)

		if req.UserID == nil {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		if err != nil {
			// log.Printf("Error checking user existence: %v", err)
			http.Error(w, "User does not exist", http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "User does not exist", http.StatusBadRequest)
			return
		}

		// Check if competition exists
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM competitions WHERE competition_id=$1)", req.CompetitionID).Scan(&exists)
		if err != nil {
			// log.Printf("Error checking stage existence: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Competition does not exist", http.StatusBadRequest)
			return
		}

		// Check if competition is open
		var isOpen int
		err = db.QueryRow("SELECT status FROM competitions WHERE competition_id=$1", req.CompetitionID).Scan(&isOpen)
		if err != nil {
			// log.Printf("Error checking competition status: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if isOpen != 1 {
			http.Error(w, "Competition is not open for signup", http.StatusBadRequest)
			return
		}

		// Check if user is already signed up for the competition
		var userSignedUp bool
		err = db.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM competition_participants WHERE competition_id=$1 AND user_id=$2)
		`, req.CompetitionID, *req.UserID).Scan(&userSignedUp)
		if err != nil {
			// log.Printf("Error checking user signup status: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if userSignedUp {
			// log.Println("User is already signed up for the competition:", *req.UserID)
			http.Error(w, "User is already signed up for the competition", http.StatusBadRequest)
			return
		}

		// check if competition is already full
		var maxParticipants int

		err = db.QueryRow(`
			SELECT max_participants FROM competitions WHERE competition_id=$1
		`, req.CompetitionID).Scan(&maxParticipants)
		if err != nil {
			// log.Printf("Error checking competition max participants: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		var numParticipants int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM competition_participants WHERE competition_id=$1
		`, req.CompetitionID).Scan(&numParticipants)
		if err != nil {
			// log.Printf("Error checking competition full status: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if numParticipants >= maxParticipants {
			// log.Println("Competition is already full:", req.CompetitionID)
			http.Error(w, "Competition is already full", http.StatusBadRequest)
			return
		}

		// Insert into stage_participants
		_, err = db.Exec(`
			INSERT INTO competition_participants (competition_id, user_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, req.CompetitionID, *req.UserID)

		if err != nil {
			// log.Printf("Error signing up: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "Signup successful"}); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
	})
}
