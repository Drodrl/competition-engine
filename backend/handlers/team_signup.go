package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type TeamSignupRequest struct {
	CompetitionID int  `json:"competition_id"`
	TeamID        *int `json:"team_id,omitempty"`
}

func NewTeamSignupHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Println("HTTP Method not allowed")
			http.Error(w, "HTTP Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TeamSignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println("Error decoding request body:", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Check if team exists
		var err error
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE team_id=$1)", *req.TeamID).Scan(&exists)
		if err != nil {
			log.Printf("Error checking team existence: %v", err)
			http.Error(w, "Team does not exist", http.StatusInternalServerError)
			return
		}
		if req.TeamID == nil {
			log.Println("Team ID is required")
			http.Error(w, "Team ID is required", http.StatusBadRequest)
			return
		}
		if !exists {
			log.Println("Team does not exist:", *req.TeamID)
			http.Error(w, "Team does not exist", http.StatusBadRequest)
			return
		}

		// Check if competition exists
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM competitions WHERE competition_id=$1)", req.CompetitionID).Scan(&exists)
		if err != nil {
			log.Printf("Error checking stage existence: %v", err)
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

		// Check if it is a team competition
		var isTeamCompetition bool
		err = db.QueryRow("SELECT flag_team FROM competitions WHERE competition_id=$1", req.CompetitionID).Scan(&isTeamCompetition)
		if err != nil {
			log.Printf("Error checking competition type: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !isTeamCompetition {
			log.Println("Competition is not a team competition:", req.CompetitionID)
			http.Error(w, "Cannot sign up to an individual competition", http.StatusBadRequest)
			return
		}

		// Check if team is already signed up for the competition
		var teamSignedUp bool
		err = db.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM competition_participants WHERE competition_id=$1 AND team_id=$2)
		`, req.CompetitionID, *req.TeamID).Scan(&teamSignedUp)
		if err != nil {
			log.Printf("Error checking team signup status: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if teamSignedUp {
			log.Println("Team is already signed up for the competition:", *req.TeamID)
			http.Error(w, "Team is already signed up for the competition", http.StatusBadRequest)
			return
		}

		// Check if any user in the team is already part of another team in the same competition
		var conflictingUsers []int
		rows, err := db.Query(`
			SELECT user_id
			FROM user_teams ut
			INNER JOIN competition_participants cp ON ut.team_id = cp.team_id
			WHERE cp.competition_id = $1 AND ut.team_id != $2
		`, req.CompetitionID, *req.TeamID)
		if err != nil {
			log.Printf("Error checking conflicting users: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		for rows.Next() {
			var userID int
			if err := rows.Scan(&userID); err != nil {
				log.Printf("Error scanning conflicting user: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			conflictingUsers = append(conflictingUsers, userID)
		}

		if len(conflictingUsers) > 0 {
			log.Printf("Conflicting users found: %v", conflictingUsers)
			http.Error(w, "Some users in the team are already part of another team in the competition", http.StatusBadRequest)
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

		// Check if athlete is a team leader
		var isTeamLeader bool
		err = db.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM user_teams WHERE team_id=$1 AND team_position='Team Leader')
		`, *req.TeamID).Scan(&isTeamLeader)
		if err != nil {
			log.Printf("Error checking team leader status: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !isTeamLeader {
			log.Println("Only team leaders can sign up for competitions")
			http.Error(w, "Only team leaders can sign up for competitions", http.StatusForbidden)
			return
		}

		// Insert into stage_participants
		_, err = db.Exec(`
			INSERT INTO competition_participants (competition_id, team_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, req.CompetitionID, *req.TeamID)

		if err != nil {
			log.Printf("Error signing up: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "Signup successful"}); err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
	})
}
