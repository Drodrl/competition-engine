package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Drodrl/competition-engine/models"
	"github.com/gorilla/mux"
)

// GET /api/public/competitions
// Optional query params: sport_id, flag_teams, status
func GetPublicCompetitions(w http.ResponseWriter, r *http.Request) {
	query := `
        SELECT competition_id, competition_name, sport_id, start_date, end_date, max_participants, organizer_id, status, date_created, date_updated, flag_teams
        FROM competitions
        WHERE status IN (1,2,3)
    `
	args := []interface{}{}
	paramCount := 1

	// Filtering by sport_id
	sportID := r.URL.Query().Get("sport_id")
	if sportID != "" {
		query += " AND sport_id = $" + strconv.Itoa(paramCount)
		args = append(args, sportID)
		paramCount++
	}
	// Filtering by flag_teams (true/false)
	flagTeams := r.URL.Query().Get("flag_teams")
	if flagTeams != "" {
		query += " AND flag_teams = $" + strconv.Itoa(paramCount)
		val, err := strconv.ParseBool(flagTeams)
		if err != nil {
			sendJSONError(w, "Invalid flag_teams value", http.StatusBadRequest)
			return
		}
		args = append(args, val)
		paramCount++
	}
	// Filtering by status
	status := r.URL.Query().Get("status")
	if status != "" {
		query += " AND status = $" + strconv.Itoa(paramCount)
		statusInt, err := strconv.Atoi(status)
		if err != nil {
			sendJSONError(w, "Invalid status value", http.StatusBadRequest)
			return
		}
		args = append(args, statusInt)
		paramCount++
	}

	_ = paramCount

	query += " ORDER BY date_created DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var competitions []models.Competition
	for rows.Next() {
		var c models.Competition
		if err := rows.Scan(&c.CompetitionId, &c.CompetitionName, &c.SportID, &c.StartDate, &c.EndDate, &c.MaxParticipants, &c.OrganizerID, &c.Status, &c.DateCreated, &c.DateUpdated, &c.FlagTeams); err != nil {
			sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		competitions = append(competitions, c)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(competitions); err != nil {
		sendJSONError(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// GET /api/public/competitions/{competitionId}/results
func GetPublicCompetitionResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["competitionId"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSONError(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}

	// Get competition status
	var status int
	if err := db.QueryRow(`SELECT status FROM competitions WHERE competition_id = $1`, id).Scan(&status); err != nil {
		sendJSONError(w, "Competition not found", http.StatusNotFound)
		return
	}
	if status != 3 {
		sendJSONError(w, "Competition is not finished yet", http.StatusBadRequest)
		return
	}

	// Get winner info
	var winnerName, teamName string
	_ = db.QueryRow(`
        SELECT u.name_user, t.team_name
        FROM matches m
        JOIN rounds r ON m.round_id = r.round_id
        JOIN match_participants mp ON mp.match_id = m.match_id
        LEFT JOIN users u ON mp.user_id = u.id_user
        LEFT JOIN teams t ON mp.team_id = t.team_id
        WHERE r.stage_id = (
            SELECT stage_id FROM competition_stages WHERE competition_id = $1 ORDER BY stage_order DESC LIMIT 1
        )
        AND mp.is_winner = true
        ORDER BY m.match_id DESC
        LIMIT 1
    `, id).Scan(&winnerName, &teamName)

	resp := map[string]interface{}{
		"competition_id": id,
		"winner": map[string]interface{}{
			"name":      winnerName,
			"team_name": teamName,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		sendJSONError(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}
