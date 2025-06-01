package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Drodrl/competition-engine/controllers"
	"github.com/Drodrl/competition-engine/models"
	"github.com/gorilla/mux"
)

// GET /api/stages/{stageId}/rounds
func GetRoundsByStageID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stageIDStr := vars["stageId"]
	stageID, err := strconv.Atoi(stageIDStr)
	if err != nil {
		sendJSONError(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}
	rows, err := db.Query(`
        SELECT round_id, stage_id, round_number 
        FROM rounds WHERE stage_id = $1 
        ORDER BY round_number
    `, stageID)
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close error: %v", err)
		}
	}()
	var rounds []models.StageRound
	for rows.Next() {
		var s models.StageRound
		if err := rows.Scan(&s.RoundID, &s.StageID, &s.RoundNumber); err != nil {
			sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rounds = append(rounds, s)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rounds); err != nil {
		log.Printf("encode error: %v", err)
		sendJSONError(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// GET /api/rounds/{roundId}/matches
func GetMatchesByRoundID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roundIDStr := vars["roundId"]
	roundID, err := strconv.Atoi(roundIDStr)
	if err != nil {
		sendJSONError(w, "Invalid round ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
        SELECT match_id, round_id, scheduled_at, completed_at
        FROM matches
        WHERE round_id = $1
    `, roundID)
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close error: %v", err)
		}
	}()

	var matches []models.Match
	for rows.Next() {
		var m models.Match
		if err := rows.Scan(&m.MatchID, &m.RoundID, &m.ScheduledAt, &m.CompletedAt); err != nil {
			sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		matches = append(matches, m)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matches); err != nil {
		log.Printf("encode error: %v", err)
		sendJSONError(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// PUT /api/matches/{matchId}/participants
func UpdateMatchResult(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchIDStr := vars["matchId"]
	matchID, err := strconv.Atoi(matchIDStr)
	if err != nil {
		sendJSONError(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	var results []struct {
		ParticipantID int  `json:"participant_id"`
		Score         int  `json:"score"`
		IsWinner      bool `json:"is_winner"`
	}
	if err := json.NewDecoder(r.Body).Decode(&results); err != nil {
		sendJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	for _, res := range results {
		_, err := db.Exec(`
            UPDATE match_participants
            SET score = $1, is_winner = $2
            WHERE match_id = $3 AND (user_id = $4 OR team_id = $4)
        `, res.Score, res.IsWinner, matchID, res.ParticipantID)
		if err != nil {
			sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// GET /api/matches/{matchId}/participants
func GetMatchParticipants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchIDStr := vars["matchId"]
	matchID, err := strconv.Atoi(matchIDStr)
	if err != nil {
		sendJSONError(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
        SELECT match_id, user_id, team_id, is_winner, score
        FROM match_participants
        WHERE match_id = $1
    `, matchID)
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close error: %v", err)
		}
	}()

	var partsList []models.MatchParticipant
	for rows.Next() {
		var p models.MatchParticipant
		if err := rows.Scan(&p.MatchID, &p.UserID, &p.TeamID, &p.IsWinner, &p.Score); err != nil {
			sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		partsList = append(partsList, p)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(partsList); err != nil {
		log.Printf("encode error: %v", err)
		sendJSONError(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// PUT /api/matches/{matchId}/results
func SaveMatchResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchIDStr := vars["matchId"]
	matchID, err := strconv.Atoi(matchIDStr)
	if err != nil {
		sendJSONError(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	var results []struct {
		ParticipantID int  `json:"participant_id"`
		Score         *int `json:"score"`
		IsWinner      bool `json:"is_winner"`
	}
	if err := json.NewDecoder(r.Body).Decode(&results); err != nil {
		sendJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rollback := func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("rollback error: %v", err)
		}
	}
	defer func() {
		if p := recover(); p != nil {
			rollback()
			panic(p)
		}
	}()

	for _, res := range results {
		_, err := tx.Exec(`
            UPDATE match_participants
            SET score = $1, is_winner = $2
            WHERE match_id = $3
              AND (user_id = $4 OR team_id = $4)
        `, res.Score, res.IsWinner, matchID, res.ParticipantID)
		if err != nil {
			rollback()
			sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	_, err = tx.Exec(`UPDATE matches SET completed_at = NOW() WHERE match_id = $1`, matchID)
	if err != nil {
		rollback()
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// POST /api/stages/{stageId}/rounds
func GenerateNextRound(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stageIDStr := vars["stageId"]
	stageID, err := strconv.Atoi(stageIDStr)
	if err != nil {
		sendJSONError(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	// Find the latest round number for this stage
	var lastRoundNumber int
	if err := db.QueryRow(`SELECT COALESCE(MAX(round_number), 0) FROM rounds WHERE stage_id = $1`, stageID).Scan(&lastRoundNumber); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if all matches in the latest round are finished
	if lastRoundNumber > 0 {
		var unfinished int
		if err := db.QueryRow(`
            SELECT COUNT(*) FROM matches
            WHERE round_id = (
                SELECT round_id FROM rounds WHERE stage_id = $1 AND round_number = $2
            ) AND completed_at IS NULL
        `, stageID, lastRoundNumber).Scan(&unfinished); err != nil {
			sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if unfinished > 0 {
			sendJSONError(w, "All matches in the current round must be finished before generating the next round.", http.StatusBadRequest)
			return
		}
	}

	// look up the stage format
	var fmtNumber int
	if err := db.QueryRow(`SELECT tourney_format_id FROM competition_stages WHERE stage_id=$1`, stageID).Scan(&fmtNumber); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	switch fmtNumber {
	case 1:
		if err := controllers.GenerateRoundSingleElim(db, stageID); err != nil {
			sendJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case 2:
		if err := controllers.GenerateRoundDoubleElim(db, stageID); err != nil {
			sendJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case 3:
		if err := controllers.GenerateRoundRobin(db, stageID); err != nil {
			sendJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		sendJSONError(w, "unsupported format", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /api/stages/{stageId}/can-generate-next-round
func CanGenerateNextRound(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stageIDStr := vars["stageId"]
	stageID, err := strconv.Atoi(stageIDStr)
	if err != nil {
		sendJSONError(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	// Get the competition_id for this stage
	var competitionID int
	if err := db.QueryRow(`SELECT competition_id FROM competition_stages WHERE stage_id = $1`, stageID).Scan(&competitionID); err != nil {
		sendJSONError(w, "Stage not found", http.StatusBadRequest)
		return
	}

	// Check competition status
	var status int
	if err := db.QueryRow(`SELECT status FROM competitions WHERE competition_id = $1`, competitionID).Scan(&status); err != nil {
		sendJSONError(w, "Competition not found", http.StatusBadRequest)
		return
	}
	if status != 2 { // 2 = Ongoing
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"canGenerate": false,
			"reason":      "Cannot generate rounds until signup is closed and competition is ongoing.",
		}); err != nil {
			log.Printf("encode error: %v", err)
		}
		return
	}

	// Find the latest round number for this stage
	var lastRoundNumber int
	if err := db.QueryRow(`SELECT COALESCE(MAX(round_number), 0) FROM rounds WHERE stage_id = $1`, stageID).Scan(&lastRoundNumber); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If no rounds, allow generation
	if lastRoundNumber == 0 {
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"canGenerate": true}); err != nil {
			log.Printf("encode error: %v", err)
		}
		return
	}

	// Check if all matches in the latest round are finished
	var unfinished int
	if err := db.QueryRow(`
        SELECT COUNT(*) FROM matches
        WHERE round_id = (
            SELECT round_id FROM rounds WHERE stage_id = $1 AND round_number = $2
        ) AND completed_at IS NULL
    `, stageID, lastRoundNumber).Scan(&unfinished); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if unfinished > 0 {
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"canGenerate": false,
			"reason":      "All matches in the current round must be finished before generating the next round.",
		}); err != nil {
			log.Printf("encode error: %v", err)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"canGenerate": true}); err != nil {
		log.Printf("encode error: %v", err)
	}
}

// POST /api/stages/{stageId}/advance
func AdvanceAfterRoundRobin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stageIDStr := vars["stageId"]
	stageID, err := strconv.Atoi(stageIDStr)
	if err != nil {
		sendJSONError(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	// 1. Check all matches completed
	var incomplete int
	if err := db.QueryRow(`
        SELECT COUNT(*) FROM matches m
        JOIN rounds r ON m.round_id = r.round_id
        WHERE r.stage_id = $1 AND m.completed_at IS NULL
    `, stageID).Scan(&incomplete); err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if incomplete > 0 {
		sendJSONError(w, "Not all matches are completed", http.StatusBadRequest)
		return
	}

	// 2. Find next stage (by stage_order)
	var nextStageID, participantsAtStart int
	err = db.QueryRow(`
        SELECT stage_id, participants_at_start FROM competition_stages
        WHERE competition_id = (SELECT competition_id FROM competition_stages WHERE stage_id = $1)
        AND stage_order = (SELECT stage_order FROM competition_stages WHERE stage_id = $1) + 1
    `, stageID).Scan(&nextStageID, &participantsAtStart)

	if err == sql.ErrNoRows {
		// No next stage: mark competition as finished
		if _, err := db.Exec(`
            UPDATE competitions SET status = 3
            WHERE competition_id = (SELECT competition_id FROM competition_stages WHERE stage_id = $1)
        `, stageID); err != nil {
			sendJSONError(w, "Failed to update competition status: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"finished":true}`)); err != nil {
			log.Printf("write error: %v", err)
		}
		return
	} else if err != nil {
		sendJSONError(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Get top N from this stage
	top, err := controllers.GetTopNFromPrevRoundRobin(db, nextStageID, participantsAtStart)
	if err != nil {
		sendJSONError(w, "Failed to get top participants: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Insert into stage_participants for next stage
	for _, e := range top {
		if e.UserID != nil {
			if _, err := db.Exec(`INSERT INTO stage_participants (stage_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, nextStageID, *e.UserID); err != nil {
				sendJSONError(w, "Failed to insert participant: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else if e.TeamID != nil {
			if _, err := db.Exec(`INSERT INTO stage_participants (stage_id, team_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, nextStageID, *e.TeamID); err != nil {
				sendJSONError(w, "Failed to insert participant: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"advanced":true}`)); err != nil {
		log.Printf("write error: %v", err)
	}
}
