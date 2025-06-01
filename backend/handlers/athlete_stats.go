package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GET /api/athletes/{userId}/stats
func GetAthleteStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		sendJSONError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// total competitions
	var totalCompetitions int
	err = db.QueryRow(`
        SELECT COUNT(DISTINCT cp.competition_id) 
        FROM competition_participants cp
        WHERE cp.user_id = $1
    `, userID).Scan(&totalCompetitions)
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	// competitions won
	var competitionsWon int
	err = db.QueryRow(`
        SELECT COUNT(DISTINCT m.match_id)
        FROM match_participants mp
        JOIN matches m ON mp.match_id = m.match_id
        JOIN rounds r ON m.round_id = r.round_id
        JOIN competition_stages cs ON r.stage_id = cs.stage_id
        JOIN competitions c ON cs.competition_id = c.competition_id
        WHERE mp.user_id = $1 AND mp.is_winner = true AND c.status = 3
    `, userID).Scan(&competitionsWon)
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	// ongoing competitions
	var ongoingCompetitions int
	err = db.QueryRow(`
        SELECT COUNT(DISTINCT c.competition_id)
        FROM competition_participants cp
        JOIN competitions c ON cp.competition_id = c.competition_id
        WHERE cp.user_id = $1 AND c.status = 2
    `, userID).Scan(&ongoingCompetitions)
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	// upcoming competitions
	var upcomingCompetitions int
	err = db.QueryRow(`
        SELECT COUNT(DISTINCT c.competition_id)
        FROM competition_participants cp
        JOIN competitions c ON cp.competition_id = c.competition_id
        WHERE cp.user_id = $1 AND c.status = 1
    `, userID).Scan(&upcomingCompetitions)
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Calculate win percentage
	winPercentage := 0
	if totalCompetitions > 0 {
		winPercentage = (competitionsWon * 100) / totalCompetitions
	}

	// Create response
	stats := struct {
		TotalCompetitions    int `json:"totalCompetitions"`
		CompetitionsWon      int `json:"competitionsWon"`
		OngoingCompetitions  int `json:"ongoingCompetitions"`
		UpcomingCompetitions int `json:"upcomingCompetitions"`
		WinPercentage        int `json:"winPercentage"`
	}{
		TotalCompetitions:    totalCompetitions,
		CompetitionsWon:      competitionsWon,
		OngoingCompetitions:  ongoingCompetitions,
		UpcomingCompetitions: upcomingCompetitions,
		WinPercentage:        winPercentage,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		sendJSONError(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// GET /api/athletes/{userId}/competitions
func GetAthleteCompetitions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		sendJSONError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
        SELECT c.competition_id, c.competition_name, s.sport_name, c.start_date, c.end_date, c.status
        FROM competition_participants cp
        JOIN competitions c ON cp.competition_id = c.competition_id
        JOIN sports s ON c.sport_id = s.sport_id
        WHERE cp.user_id = $1
        ORDER BY 
            CASE 
                WHEN c.status = 2 THEN 1 -- Ongoing first
                WHEN c.status = 1 THEN 2 -- Upcoming second
                WHEN c.status = 3 THEN 3 -- Completed last
                ELSE 4 -- Other status
            END,
            c.start_date DESC
    `, userID)
	if err != nil {
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Competition struct {
		CompetitionID   int     `json:"competition_id"`
		CompetitionName string  `json:"competition_name"`
		SportName       string  `json:"sport_name"`
		StartDate       string  `json:"start_date"`
		EndDate         *string `json:"end_date"`
		Status          int     `json:"status"`
	}

	var competitions []Competition
	for rows.Next() {
		var c Competition
		err := rows.Scan(&c.CompetitionID, &c.CompetitionName, &c.SportName, &c.StartDate, &c.EndDate, &c.Status)
		if err != nil {
			sendJSONError(w, "Error scanning competition data", http.StatusInternalServerError)
			return
		}
		competitions = append(competitions, c)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(competitions); err != nil {
		sendJSONError(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
