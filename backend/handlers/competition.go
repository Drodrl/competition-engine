package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Drodrl/competition-engine/models"
	"github.com/gorilla/mux"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

// POST /api/competitions/draft
func CreateDraftCompetition(w http.ResponseWriter, r *http.Request) {
	var competition models.Competition
	if err := json.NewDecoder(r.Body).Decode(&competition); err != nil {
		sendJSONError(w, "Invalid JSON", http.StatusBadRequest)
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
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]int{"competition_id": competitionID}); err != nil {
		sendJSONError(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// DELETE /api/competitions/{competitionId}
func DeleteCompetition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["competitionId"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSONError(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	if _, err = db.Exec(`DELETE FROM competitions WHERE competition_id = $1`, id); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err = db.Exec(`DELETE FROM competition_stages WHERE competition_id = $1`, id); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err = db.Exec(`DELETE FROM competition_participants WHERE competition_id = $1`, id); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/competitions/organizer/{organizerId}
func GetCompetitionsByOrganizer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizerIDStr := vars["organizerId"]
	organizerID, err := strconv.Atoi(organizerIDStr)
	if err != nil {
		sendJSONError(w, "Invalid organizer ID", http.StatusBadRequest)
		return
	}
	rows, err := db.Query(`
        SELECT c.competition_id, c.competition_name, c.sport_id, c.start_date, c.end_date, c.max_participants, c.organizer_id, c.status, c.date_created, c.date_updated, c.flag_teams, s.sport_name
        FROM competitions c
        JOIN sports s ON c.sport_id = s.sport_id
        WHERE c.organizer_id = $1
        ORDER BY c.date_created DESC
    `, organizerID)
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close error: %v", err)
		}
	}()
	var competitions []models.Competition
	for rows.Next() {
		var c models.Competition
		if err := rows.Scan(&c.CompetitionId, &c.CompetitionName, &c.SportID, &c.StartDate, &c.EndDate, &c.MaxParticipants, &c.OrganizerID, &c.Status, &c.DateCreated, &c.DateUpdated, &c.FlagTeams, &c.SportName); err != nil {
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

// GET /api/competitions/{competitionId}
func GetCompetitionByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["competitionId"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSONError(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	var c models.Competition
	err = db.QueryRow(`
        SELECT c.competition_id, c.competition_name, c.sport_id, c.start_date, c.end_date, c.max_participants, c.organizer_id, c.status, c.date_created, c.date_updated, c.flag_teams, s.sport_name
        FROM competitions c
        JOIN sports s ON c.sport_id = s.sport_id
        WHERE c.competition_id = $1
    `, id).Scan(&c.CompetitionId, &c.CompetitionName, &c.SportID, &c.StartDate, &c.EndDate, &c.MaxParticipants, &c.OrganizerID, &c.Status, &c.DateCreated, &c.DateUpdated, &c.FlagTeams, &c.SportName)
	if err != nil {
		sendJSONError(w, "Competition not found", http.StatusNotFound)
		return
	}

	resp := map[string]interface{}{
		"competition_id":   c.CompetitionId,
		"competition_name": c.CompetitionName,
		"sport_id":         c.SportID,
		"sport_name":       c.SportName,
		"start_date":       c.StartDate,
		"end_date":         c.EndDate,
		"max_participants": c.MaxParticipants,
		"organizer_id":     c.OrganizerID,
		"status":           c.Status,
		"date_created":     c.DateCreated,
		"date_updated":     c.DateUpdated,
		"flag_teams":       c.FlagTeams,
	}

	if c.Status == 3 {
		// Competition finished, fetch winner
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
        `, c.CompetitionId).Scan(&winnerName, &teamName)
		resp["winner"] = map[string]interface{}{
			"name":      winnerName,
			"team_name": teamName,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("encode error: %v", err)
	}
}

// PUT /api/competitions/{competitionId}
func UpdateCompetition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["competitionId"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSONError(w, "Invalid competition ID", http.StatusBadRequest)
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
		sendJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if _, err = db.Exec(`
        UPDATE competitions
        SET competition_name = $1, start_date = $2, end_date = $3, max_participants = $4, flag_teams = $5, date_updated = $6
        WHERE competition_id = $7
    `, req.CompetitionName, req.StartDate, req.EndDate, req.MaxParticipants, req.FlagTeams, time.Now(), id); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// PATCH /api/competitions/{competitionId}/status
func ChangeCompetitionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["competitionId"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSONError(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	var currentStatus, maxParticipants, sportID int
	err = db.QueryRow(`SELECT status, max_participants, sport_id FROM competitions WHERE competition_id = $1`, id).Scan(&currentStatus, &maxParticipants, &sportID)
	if err != nil {
		sendJSONError(w, "Competition not found", http.StatusNotFound)
		return
	}
	if currentStatus == req.Status {
		sendJSONError(w, "Competition already in this status", http.StatusBadRequest)
		return
	}

	// Check requirements for opening or closing
	if req.Status == 1 {
		if err := canOpenCompetition(id, maxParticipants); err != nil {
			sendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if req.Status == 2 {
		if err := canCloseSignup(id); err != nil {
			sendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Insert all participants into the first stage
		var firstStageID int
		err = db.QueryRow(`
            SELECT stage_id FROM competition_stages WHERE competition_id = $1 ORDER BY stage_order ASC LIMIT 1
        `, id).Scan(&firstStageID)
		if err != nil {
			sendJSONError(w, "No stages found for competition", http.StatusBadRequest)
			return
		}
		// Insert users
		res, err := db.Exec(`
            INSERT INTO stage_participants (stage_id, user_id)
            SELECT $1, user_id FROM competition_participants WHERE competition_id = $2 AND user_id IS NOT NULL
            ON CONFLICT DO NOTHING
        `, firstStageID, id)
		if err != nil {
			sendJSONError(w, "Failed to insert users into stage_participants: "+err.Error(), http.StatusInternalServerError)
			return
		}
		count, _ := res.RowsAffected()
		log.Printf("Inserted %d user participants into stage_participants", count)
		// Insert teams
		if _, err = db.Exec(`
            INSERT INTO stage_participants (stage_id, team_id)
            SELECT $1, team_id FROM competition_participants WHERE competition_id = $2 AND team_id IS NOT NULL
            ON CONFLICT DO NOTHING
        `, firstStageID, id); err != nil {
			sendJSONError(w, "Failed to insert teams into stage_participants: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if _, err = db.Exec(`UPDATE competitions SET status = $1, date_updated = $2 WHERE competition_id = $3`, req.Status, time.Now(), id); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Helper: Check if competition can be opened (status 1)
func canOpenCompetition(competitionID int, maxParticipants int) error {
	stages, err := getCompetitionStages(competitionID)
	if err != nil {
		return err
	}
	if len(stages) == 0 {
		return errors.New("at least one stage is required to open the competition")
	}
	formatMin, err := getTournamentFormatMinimums()
	if err != nil {
		return err
	}
	prevParticipants := maxParticipants
	for i, s := range stages {
		minimum := formatMin[s.TourneyFormatID]
		if s.ParticipantsAtStart < minimum {
			return errors.New("Stage '" + s.StageName + "' requires at least " + strconv.Itoa(minimum) + " participants.")
		}
		if s.ParticipantsAtStart%2 != 0 {
			return errors.New("Stage '" + s.StageName + "' must have an even number of participants at start.")
		}
		if i > 0 && s.ParticipantsAtStart > prevParticipants-2 {
			return errors.New("Stage '" + s.StageName + "' cannot have more participants at start than previous stage's end minus 2 (" + strconv.Itoa(prevParticipants-2) + ").")
		}
		prevParticipants = s.ParticipantsAtStart
	}
	return nil
}

// Helper: Get all stages for a competition
func getCompetitionStages(competitionID int) ([]models.StageDTO, error) {
	rows, err := db.Query(`
        SELECT stage_id, stage_name, stage_order, tourney_format_id, participants_at_start, participants_at_end
        FROM competition_stages WHERE competition_id = $1 ORDER BY stage_order ASC
    `, competitionID)
	if err != nil {
		return nil, errors.New("DB error: " + err.Error())
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close error: %v", err)
		}
	}()
	var stages []models.StageDTO
	for rows.Next() {
		var s models.StageDTO
		if err := rows.Scan(&s.StageID, &s.StageName, &s.StageOrder, &s.TourneyFormatID, &s.ParticipantsAtStart, &s.ParticipantsAtEnd); err != nil {
			return nil, errors.New("DB error: " + err.Error())
		}
		stages = append(stages, s)
	}
	return stages, nil
}

// Helper: Get minimum participants for each tournament format
func getTournamentFormatMinimums() (map[int]int, error) {
	rows, err := db.Query(`SELECT tourney_format_id, min_participants FROM tournament_formats`)
	if err != nil {
		return nil, errors.New("DB error: " + err.Error())
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close error: %v", err)
		}
	}()
	formatMin := make(map[int]int)
	for rows.Next() {
		var id, min int
		if err := rows.Scan(&id, &min); err != nil {
			return nil, errors.New("DB error: " + err.Error())
		}
		formatMin[id] = min
	}
	return formatMin, nil
}

func getCompetitionMaxParticipants(competitionID int) (int, error) {
	var maxParticipants int
	if err := db.QueryRow(`SELECT max_participants FROM competitions WHERE competition_id = $1`, competitionID).Scan(&maxParticipants); err != nil {
		return 0, errors.New("DB error: " + err.Error())
	}
	return maxParticipants, nil
}

// Helper: Check if signup can be closed (status 2)
func canCloseSignup(competitionID int) error {
	maxParticipants, err := getCompetitionMaxParticipants(competitionID)
	if err != nil {
		return err
	}
	var numParticipants int
	if err := db.QueryRow(`SELECT COUNT(*) FROM competition_participants WHERE competition_id = $1`, competitionID).Scan(&numParticipants); err != nil {
		return errors.New("DB error: " + err.Error())
	}
	if numParticipants < maxParticipants {
		return errors.New("not enough participants to close signup")
	}
	return nil
}

// Multiplexer for /api/competitions/{competitionId} and /api/competitions/{competitionId}/status
func CompetitionByIDHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			ChangeCompetitionStatus(w, r)
		} else if r.Method == http.MethodGet {
			GetCompetitionByID(w, r)
		} else if r.Method == http.MethodPut {
			UpdateCompetition(w, r)
		} else if r.Method == http.MethodDelete {
			DeleteCompetition(w, r)
		} else {
			sendJSONError(w, "Not found", http.StatusNotFound)
		}
	})
}

// GET /api/competitions/{competitionId}/stages
func GetStagesByCompetitionID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	competitionIDStr := vars["competitionId"]
	competitionID, err := strconv.Atoi(competitionIDStr)
	if err != nil {
		sendJSONError(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	rows, err := db.Query(`
        SELECT stage_id, stage_name, stage_order, tourney_format_id, participants_at_start, participants_at_end
        FROM competition_stages WHERE competition_id = $1 ORDER BY stage_order ASC
    `, competitionID)
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close error: %v", err)
		}
	}()
	var stages []models.StageDTO
	for rows.Next() {
		var s models.StageDTO
		if err := rows.Scan(&s.StageID, &s.StageName, &s.StageOrder, &s.TourneyFormatID, &s.ParticipantsAtStart, &s.ParticipantsAtEnd); err != nil {
			sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		stages = append(stages, s)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stages); err != nil {
		sendJSONError(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// POST /api/competitions/{competitionId}/stages
func AddStageToCompetition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	competitionIDStr := vars["competitionId"]
	competitionID, err := strconv.Atoi(competitionIDStr)
	if err != nil {
		sendJSONError(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	var stage models.StageDTO
	if err := json.NewDecoder(r.Body).Decode(&stage); err != nil {
		sendJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	stages, err := getCompetitionStages(competitionID)
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	stages = append(stages, stage)
	maxParticipants, _ := getCompetitionMaxParticipants(competitionID)
	if err := validateStagesBusinessRules(competitionID, stages, maxParticipants); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err = db.Exec(`
        INSERT INTO competition_stages (competition_id, stage_order, stage_name, tourney_format_id, participants_at_start, participants_at_end)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, competitionID, stage.StageOrder, stage.StageName, stage.TourneyFormatID, stage.ParticipantsAtStart, stage.ParticipantsAtEnd); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// PUT /api/competitions/{competitionId}/stages/{stageId}
func UpdateStage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	competitionIDStr := vars["competitionId"]
	stageIDStr := vars["stageId"]

	competitionID, err := strconv.Atoi(competitionIDStr)
	if err != nil {
		sendJSONError(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}
	stageID, err := strconv.Atoi(stageIDStr)
	if err != nil {
		sendJSONError(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	var stage models.StageDTO
	if err := json.NewDecoder(r.Body).Decode(&stage); err != nil {
		sendJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	stages, err := getCompetitionStages(competitionID)
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range stages {
		if stages[i].StageID == stageID {
			stage.StageID = stageID
			stages[i] = stage
			break
		}
	}

	maxParticipants, _ := getCompetitionMaxParticipants(competitionID)
	if err := validateStagesBusinessRules(competitionID, stages, maxParticipants); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err = db.Exec(`
        UPDATE competition_stages
        SET stage_name = $1, stage_order = $2, tourney_format_id = $3, participants_at_start = $4, participants_at_end = $5
        WHERE stage_id = $6
    `, stage.StageName, stage.StageOrder, stage.TourneyFormatID, stage.ParticipantsAtStart, stage.ParticipantsAtEnd, stageID); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DELETE /api/competitions/{competitionId}/stages/{stageId}
func DeleteStage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stageIDStr := vars["stageId"]
	stageID, err := strconv.Atoi(stageIDStr)
	if err != nil {
		sendJSONError(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}
	if _, err = db.Exec(`DELETE FROM competition_stages WHERE stage_id = $1`, stageID); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /api/competitions/{competitionId}/participants
func GetParticipantsByCompetitionID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	competitionIDStr := vars["competitionId"]
	competitionID, err := strconv.Atoi(competitionIDStr)
	if err != nil {
		sendJSONError(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
        SELECT u.id_user, u.name_user, u.lname1_user
        FROM competition_participants cp
        JOIN users u ON cp.user_id = u.id_user
        WHERE cp.competition_id = $1
        UNION
        SELECT t.team_id, t.team_name, NULL
        FROM competition_participants cp
        JOIN teams t ON cp.team_id = t.team_id
        WHERE cp.competition_id = $1
    `, competitionID)
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close error: %v", err)
		}
	}()

	type Participant struct {
		ID       int     `json:"id"`
		Name     string  `json:"name"`
		LastName *string `json:"last_name,omitempty"`
	}
	var participants []Participant
	for rows.Next() {
		var p Participant
		if err := rows.Scan(&p.ID, &p.Name, &p.LastName); err != nil {
			sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		participants = append(participants, p)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(participants); err != nil {
		sendJSONError(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// POST /api/competitions/{competitionId}/finish
func FinishCompetition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	competitionIDStr := vars["competitionId"]
	competitionID, err := strconv.Atoi(competitionIDStr)
	if err != nil {
		sendJSONError(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}

	// 1. Find last stage
	var lastStageID int
	err = db.QueryRow(`
        SELECT stage_id FROM competition_stages
        WHERE competition_id = $1
        ORDER BY stage_order DESC LIMIT 1
    `, competitionID).Scan(&lastStageID)
	if err != nil {
		sendJSONError(w, "No stages found for competition", http.StatusBadRequest)
		return
	}

	// 2. Check all matches in last stage are finished
	var unfinished int
	if err = db.QueryRow(`
        SELECT COUNT(*) FROM matches m
        JOIN rounds r ON m.round_id = r.round_id
        WHERE r.stage_id = $1 AND m.completed_at IS NULL
    `, lastStageID).Scan(&unfinished); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if unfinished > 0 {
		sendJSONError(w, "Not all matches are completed", http.StatusBadRequest)
		return
	}

	// 3. Find the last round in the last stage
	var lastRoundID int
	if err = db.QueryRow(`
        SELECT round_id FROM rounds
        WHERE stage_id = $1
        ORDER BY round_number DESC LIMIT 1
    `, lastStageID).Scan(&lastRoundID); err != nil {
		sendJSONError(w, "No rounds found in last stage", http.StatusBadRequest)
		return
	}

	// 4. Find the winner(s) in the last match of the last round
	var winnerUserID, winnerTeamID sql.NullInt64
	if err = db.QueryRow(`
        SELECT mp.user_id, mp.team_id
        FROM matches m
        JOIN match_participants mp ON mp.match_id = m.match_id
        WHERE m.round_id = $1 AND mp.is_winner = true
        LIMIT 1
    `, lastRoundID).Scan(&winnerUserID, &winnerTeamID); err != nil {
		sendJSONError(w, "No winner found in last round", http.StatusBadRequest)
		return
	}

	var winnerName, teamName string
	if winnerUserID.Valid {
		_ = db.QueryRow(`SELECT name_user FROM users WHERE id_user = $1`, winnerUserID.Int64).Scan(&winnerName)
	}
	if winnerTeamID.Valid {
		_ = db.QueryRow(`SELECT team_name FROM teams WHERE team_id = $1`, winnerTeamID.Int64).Scan(&teamName)
	}

	// 5. Update competition status
	if _, err = db.Exec(`UPDATE competitions SET status = 3 WHERE competition_id = $1`, competitionID); err != nil {
		sendJSONError(w, "Failed to update competition status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Return winner info in response
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"finished": true,
		"winner": map[string]interface{}{
			"name":      winnerName,
			"team_name": teamName,
		},
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("encode error: %v", err)
	}
}

// GET /api/competitions
func GetAllCompetitions(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
        SELECT competition_id, competition_name, sport_id, start_date, end_date, max_participants, organizer_id, status, date_created, date_updated, flag_teams
        FROM competitions
        ORDER BY date_created DESC
    `)
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close error: %v", err)
		}
	}()
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

// GET /api/competitions/flag_teams/{flagTeams}
func GetCompetitionsByFlagTeams(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flagTeams := vars["flagTeams"]
	isTeamCompetition, err := strconv.ParseBool(flagTeams)
	if err != nil {
		sendJSONError(w, "Invalid flag teams value", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
        SELECT competition_id, competition_name, sport_id, start_date, end_date, max_participants, organizer_id, status, date_created, date_updated, flag_teams
        FROM competitions
        WHERE flag_teams = $1
        ORDER BY date_created DESC
    `, isTeamCompetition)
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close error: %v", err)
		}
	}()

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

// Helper: Validate stage business rules
func validateStagesBusinessRules(competitionID int, stages []models.StageDTO, maxParticipants int) error {
	if len(stages) == 0 {
		return errors.New("please add at least one stage before saving")
	}
	for i, stage := range stages {
		// 1st stage, only one stage: must be single or double elim
		if i == 0 && len(stages) == 1 && !(stage.TourneyFormatID == 1 || stage.TourneyFormatID == 2) {
			return errors.New("if there is only one stage, it must be Single or Double Elimination")
		}
		// 1st stage, two stages: must be RR
		if i == 0 && len(stages) == 2 && stage.TourneyFormatID != 3 {
			return errors.New("if there are two stages, the first must be Round Robin")
		}
		// 2nd stage, two stages: must be single or double elim
		if i == 1 && len(stages) == 2 && !(stage.TourneyFormatID == 1 || stage.TourneyFormatID == 2) {
			return errors.New("the last stage must be Single or Double Elimination")
		}
		// Minimum participants
		minParticipants, err := getFormatMinParticipants(stage.TourneyFormatID)
		if err != nil {
			return err
		}
		if stage.ParticipantsAtStart < minParticipants {
			return errors.New("stage requires at least the minimum number of participants")
		}
		// Even number check
		if stage.ParticipantsAtStart%2 != 0 {
			return errors.New("participants at start must be an even number")
		}
		// Participants at start for subsequent stages
		if i > 0 {
			prev := stages[i-1]
			if stage.ParticipantsAtStart > prev.ParticipantsAtStart-2 {
				return errors.New("participants at start cannot exceed participants at start of previous stage (at least 2 less)")
			}
		}
	}
	return nil
}

func getFormatMinParticipants(formatID int) (int, error) {
	var min int
	if err := db.QueryRow(`SELECT min_participants FROM tournament_formats WHERE tourney_format_id = $1`, formatID).Scan(&min); err != nil {
		return 0, err
	}
	return min, nil
}
