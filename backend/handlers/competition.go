package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Drodrl/competition-engine/models"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

// POST /api/competitions/draft
func CreateDraftCompetition(w http.ResponseWriter, r *http.Request) {
	var competition models.Competition
	if err := json.NewDecoder(r.Body).Decode(&competition); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	now := time.Now()
	var competitionID int
	err := db.QueryRow(`
        INSERT INTO competitions
            (competition_name, sport_id, start_date, end_date, organizer_id, status, date_created, date_updated, max_participants, flag_teams)
        VALUES ($1,$2,$3,$4,$5,0,$6,$6,$7,$8)
        RETURNING competition_id
    `, competition.CompetitionName, competition.SportID, competition.StartDate, competition.EndDate, competition.OrganizerID, now, competition.MaxParticipants, competition.FlagTeams).Scan(&competitionID)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]int{"competition_id": competitionID}); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// DELETE /api/competitions/{id}
func DeleteCompetition(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Competition ID required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	_, err = db.Exec(`DELETE FROM competitions WHERE competition_id = $1`, id)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = db.Exec(`DELETE FROM competition_stages WHERE competition_id = $1`, id)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = db.Exec(`DELETE FROM competition_participants WHERE competition_id = $1`, id)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/competitions/organizer/{organizerId}
func GetCompetitionsByOrganizer(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Organizer ID required", http.StatusBadRequest)
		return
	}
	organizerID, err := strconv.Atoi(parts[4])
	if err != nil {
		http.Error(w, "Invalid organizer ID", http.StatusBadRequest)
		return
	}
	rows, err := db.Query(`
		SELECT competition_id, competition_name, sport_id, start_date, end_date, max_participants, organizer_id, status, date_created, date_updated, flag_teams
		FROM competitions WHERE organizer_id = $1
		ORDER BY date_created DESC
	`, organizerID)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var competitions []models.Competition
	for rows.Next() {
		var c models.Competition
		err := rows.Scan(&c.CompetitionId, &c.CompetitionName, &c.SportID, &c.StartDate, &c.EndDate, &c.MaxParticipants, &c.OrganizerID, &c.Status, &c.DateCreated, &c.DateUpdated, &c.FlagTeams)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		competitions = append(competitions, c)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(competitions); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// GET /api/competitions/{id}
func GetCompetitionByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Competition ID required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	var c models.Competition
	err = db.QueryRow(`
        SELECT competition_id, competition_name, sport_id, start_date, end_date, max_participants, organizer_id, status, date_created, date_updated, flag_teams
        FROM competitions WHERE competition_id = $1
    `, id).Scan(&c.CompetitionId, &c.CompetitionName, &c.SportID, &c.StartDate, &c.EndDate, &c.MaxParticipants, &c.OrganizerID, &c.Status, &c.DateCreated, &c.DateUpdated, &c.FlagTeams)
	if err != nil {
		http.Error(w, "Competition not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(c); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// PUT /api/competitions/{id}
func UpdateCompetition(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Competition ID required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	var req struct {
		CompetitionName string  `json:"competition_name"`
		StartDate       *string `json:"start_date"`
		EndDate         *string `json:"end_date"`
		MaxParticipants *int    `json:"max_participants"`
		FlagTeams       bool    `json:"flag_teams"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	_, err = db.Exec(`
        UPDATE competitions
        SET competition_name = $1, start_date = $2, end_date = $3, max_participants = $4, flag_teams = $5, date_updated = $6
        WHERE competition_id = $7
    `, req.CompetitionName, req.StartDate, req.EndDate, req.MaxParticipants, req.FlagTeams, time.Now(), id)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// PATCH /api/competitions/{id}/status
func ChangeCompetitionStatus(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Competition ID required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	var currentStatus int
	err = db.QueryRow(`SELECT status FROM competitions WHERE competition_id = $1`, id).Scan(&currentStatus)
	if err != nil {
		http.Error(w, "Competition not found", http.StatusNotFound)
		return
	}
	if currentStatus == req.Status {
		http.Error(w, "Competition already in this status", http.StatusBadRequest)
		return
	}
	_, err = db.Exec(`UPDATE competitions SET status = $1, date_updated = $2 WHERE competition_id = $3`, req.Status, time.Now(), id)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Multiplexer for /api/competitions/{id} and /api/competitions/{id}/status
func CompetitionByIDHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/status") && r.Method == http.MethodPatch {
			ChangeCompetitionStatus(w, r)
		} else if r.Method == http.MethodGet {
			GetCompetitionByID(w, r)
		} else if r.Method == http.MethodPut {
			UpdateCompetition(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
}

// GET /api/competitions/{id}/stages
func GetStagesByCompetitionID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Competition ID required", http.StatusBadRequest)
		return
	}
	competitionID, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	rows, err := db.Query(`
        SELECT stage_id, stage_name, stage_order, tourney_format_id, participants_at_start, participants_at_end
        FROM competition_stages WHERE competition_id = $1 ORDER BY stage_order ASC
    `, competitionID)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var stages []models.StageDTO
	for rows.Next() {
		var s models.StageDTO
		if err := rows.Scan(&s.StageID, &s.StageName, &s.StageOrder, &s.TourneyFormatID, &s.ParticipantsAtStart, &s.ParticipantsAtEnd); err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		stages = append(stages, s)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stages); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// POST /api/competitions/{id}/stages
func AddStageToCompetition(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Competition ID required", http.StatusBadRequest)
		return
	}
	competitionID, err := strconv.Atoi(parts[3])
	if err != nil {
		http.Error(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	var stage models.StageDTO
	if err := json.NewDecoder(r.Body).Decode(&stage); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	_, err = db.Exec(`
        INSERT INTO competition_stages (competition_id, stage_order, stage_name, tourney_format_id, participants_at_start, participants_at_end)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, competitionID, stage.StageOrder, stage.StageName, stage.TourneyFormatID, stage.ParticipantsAtStart, stage.ParticipantsAtEnd)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// PUT /api/competitions/{id}/stages/{stageId}
func UpdateStage(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 7 {
		http.Error(w, "Stage ID required", http.StatusBadRequest)
		return
	}
	stageID, err := strconv.Atoi(parts[5])
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}
	var stage models.StageDTO
	if err := json.NewDecoder(r.Body).Decode(&stage); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	_, err = db.Exec(`
        UPDATE competition_stages
        SET stage_name = $1, stage_order = $2, tourney_format_id = $3,  participants_at_start = $4, participants_at_end = $5
        WHERE stage_id = $6
    `, stage.StageName, stage.StageOrder, stage.TourneyFormatID, stage.ParticipantsAtStart, stage.ParticipantsAtEnd, stageID)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DELETE /api/competitions/{id}/stages/{stageId}
func DeleteStage(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 7 {
		http.Error(w, "Stage ID required", http.StatusBadRequest)
		return
	}
	stageID, err := strconv.Atoi(parts[5])
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}
	_, err = db.Exec(`DELETE FROM competition_stages WHERE stage_id = $1`, stageID)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /api/competitions
func GetAllCompetitions(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
        SELECT competition_id, competition_name, sport_id, start_date, end_date, max_participants, organizer_id, status, date_created, date_updated, flag_teams
        FROM competitions
        ORDER BY date_created DESC
    `)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var competitions []models.Competition
	for rows.Next() {
		var c models.Competition
		err := rows.Scan(&c.CompetitionId, &c.CompetitionName, &c.SportID, &c.StartDate, &c.EndDate, &c.MaxParticipants, &c.OrganizerID, &c.Status, &c.DateCreated, &c.DateUpdated, &c.FlagTeams)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		competitions = append(competitions, c)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(competitions); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// GET /api/competitions/flag_teams/{flagTeams}
func GetCompetitionsByFlagTeams(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Flag teams value required", http.StatusBadRequest)
		return
	}

	flagTeams := parts[4]
	isTeamCompetition, err := strconv.ParseBool(flagTeams)
	if err != nil {
		http.Error(w, "Invalid flag teams value", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
        SELECT competition_id, competition_name, sport_id, start_date, end_date, max_participants, organizer_id, status, date_created, date_updated, flag_teams
        FROM competitions
        WHERE flag_teams = $1
        ORDER BY date_created DESC
    `, isTeamCompetition)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var competitions []models.Competition
	for rows.Next() {
		var c models.Competition
		err := rows.Scan(&c.CompetitionId, &c.CompetitionName, &c.SportID, &c.StartDate, &c.EndDate, &c.MaxParticipants, &c.OrganizerID, &c.Status, &c.DateCreated, &c.DateUpdated, &c.FlagTeams)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		competitions = append(competitions, c)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(competitions); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}
