package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type CreateTeamRequest struct {
	TeamName string `json:"teamName"`
	UserIDs  []int  `json:"userIds"`
}

func NewTeamCreateHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateTeamRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println("Error decoding request body:", err)
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if req.TeamName == "" {
			log.Println("Team name is required")
			http.Error(w, "Team name is required", http.StatusBadRequest)
			return
		}

		// check if team name already exists
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE team_name=$1)", req.TeamName).Scan(&exists)
		if err != nil {
			log.Printf("Error checking team name existence: %v", err)
			http.Error(w, "Failed to check team name", http.StatusInternalServerError)
			return
		}

		if exists {
			log.Println("Team name already exists:", req.TeamName)
			http.Error(w, "Team name already exists", http.StatusBadRequest)
			return
		}

		// Check if user IDs are provided
		if len(req.UserIDs) == 0 {
			log.Println("At least one user ID is required to create a team")
			http.Error(w, "At least one user ID is required", http.StatusBadRequest)
			return
		}

		// Insert the team into the teams table
		dateCreated := time.Now()
		var teamID int
		teamPosition := "Team Leader" // Default position for the team creator
		err = db.QueryRow("INSERT INTO teams (team_name, date_created, date_updated) VALUES ($1, $2, $3) RETURNING team_id", req.TeamName, dateCreated, dateCreated).Scan(&teamID)
		if err != nil {
			log.Printf("Failed to create team: %v", err)
			http.Error(w, "Failed to create team", http.StatusInternalServerError)
			return
		}

		// Insert the team members into the user_teams table
		for _, userID := range req.UserIDs {
			_, err := db.Exec("INSERT INTO user_teams (team_id, user_id, team_position, date_created, date_updated) VALUES ($1, $2, $3, $4, $5)", teamID, userID, teamPosition, dateCreated, dateCreated)
			if err != nil {
				log.Printf("Failed to add user %d to team %d: %v", userID, teamID, err)
				http.Error(w, "Failed to add team members", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
	})
}
