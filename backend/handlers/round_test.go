package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Drodrl/competition-engine/models"
)

func TestGetRoundsByStageID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT round_id, stage_id, round_number").
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"round_id", "stage_id", "round_number"}).
			AddRow(1, 5, 1).
			AddRow(2, 5, 2))
	req := httptest.NewRequest(http.MethodGet, "/api/stages/5/rounds", nil)
	req = muxSetVars(req, map[string]string{"stageId": "5"})
	rr := httptest.NewRecorder()
	GetRoundsByStageID(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var rounds []models.StageRound
	if err := json.NewDecoder(rr.Body).Decode(&rounds); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(rounds) != 2 || rounds[0].RoundID != 1 || rounds[1].RoundID != 2 {
		t.Errorf("unexpected rounds: %+v", rounds)
	}
}

func TestGetRoundsByStageID_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodGet, "/api/stages/abc/rounds", nil)
	req = muxSetVars(req, map[string]string{"stageId": "abc"})
	rr := httptest.NewRecorder()
	GetRoundsByStageID(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestGetRoundsByStageID_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT round_id, stage_id, round_number").
		WithArgs(5).
		WillReturnError(errors.New("db fail"))
	req := httptest.NewRequest(http.MethodGet, "/api/stages/5/rounds", nil)
	req = muxSetVars(req, map[string]string{"stageId": "5"})
	rr := httptest.NewRecorder()
	GetRoundsByStageID(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 InternalServerError, got %d", rr.Code)
	}
}

func TestGetMatchesByRoundID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT match_id, round_id, scheduled_at, completed_at").
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"match_id", "round_id", "scheduled_at", "completed_at"}).
			AddRow(1, 3, nil, nil))
	req := httptest.NewRequest(http.MethodGet, "/api/rounds/3/matches", nil)
	req = muxSetVars(req, map[string]string{"roundId": "3"})
	rr := httptest.NewRecorder()
	GetMatchesByRoundID(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var matches []models.Match
	if err := json.NewDecoder(rr.Body).Decode(&matches); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(matches) != 1 || matches[0].MatchID != 1 {
		t.Errorf("unexpected matches: %+v", matches)
	}
}

func TestGetMatchesByRoundID_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodGet, "/api/rounds/abc/matches", nil)
	req = muxSetVars(req, map[string]string{"roundId": "abc"})
	rr := httptest.NewRecorder()
	GetMatchesByRoundID(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestGetMatchesByRoundID_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT match_id, round_id, scheduled_at, completed_at").
		WithArgs(3).
		WillReturnError(errors.New("db fail"))
	req := httptest.NewRequest(http.MethodGet, "/api/rounds/3/matches", nil)
	req = muxSetVars(req, map[string]string{"roundId": "3"})
	rr := httptest.NewRecorder()
	GetMatchesByRoundID(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 InternalServerError, got %d", rr.Code)
	}
}

func TestUpdateMatchResult_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectExec("UPDATE match_participants").
		WithArgs(10, true, 2, 5).
		WillReturnResult(sqlmock.NewResult(1, 1))
	body := `[{"participant_id":5,"score":10,"is_winner":true}]`
	req := httptest.NewRequest(http.MethodPut, "/api/matches/2/participants", bytes.NewReader([]byte(body)))
	req = muxSetVars(req, map[string]string{"matchId": "2"})
	rr := httptest.NewRecorder()
	UpdateMatchResult(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestUpdateMatchResult_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPut, "/api/matches/abc/participants", nil)
	req = muxSetVars(req, map[string]string{"matchId": "abc"})
	rr := httptest.NewRecorder()
	UpdateMatchResult(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestUpdateMatchResult_BadJSON(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPut, "/api/matches/2/participants", bytes.NewReader([]byte("bad json")))
	req = muxSetVars(req, map[string]string{"matchId": "2"})
	rr := httptest.NewRecorder()
	UpdateMatchResult(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestUpdateMatchResult_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectExec("UPDATE match_participants").
		WithArgs(10, true, 2, 5).
		WillReturnError(errors.New("db fail"))
	body := `[{"participant_id":5,"score":10,"is_winner":true}]`
	req := httptest.NewRequest(http.MethodPut, "/api/matches/2/participants", bytes.NewReader([]byte(body)))
	req = muxSetVars(req, map[string]string{"matchId": "2"})
	rr := httptest.NewRecorder()
	UpdateMatchResult(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 InternalServerError, got %d", rr.Code)
	}
}

func TestGetMatchParticipants_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT match_id, user_id, team_id, is_winner, score").
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"match_id", "user_id", "team_id", "is_winner", "score"}).
			AddRow(3, 1, nil, true, 10))
	req := httptest.NewRequest(http.MethodGet, "/api/matches/3/participants", nil)
	req = muxSetVars(req, map[string]string{"matchId": "3"})
	rr := httptest.NewRecorder()
	GetMatchParticipants(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var parts []models.MatchParticipant
	if err := json.NewDecoder(rr.Body).Decode(&parts); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(parts) != 1 || parts[0].MatchID != 3 {
		t.Errorf("unexpected participants: %+v", parts)
	}
}

func TestGetMatchParticipants_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodGet, "/api/matches/abc/participants", nil)
	req = muxSetVars(req, map[string]string{"matchId": "abc"})
	rr := httptest.NewRecorder()
	GetMatchParticipants(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestGetMatchParticipants_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT match_id, user_id, team_id, is_winner, score").
		WithArgs(3).
		WillReturnError(errors.New("db fail"))
	req := httptest.NewRequest(http.MethodGet, "/api/matches/3/participants", nil)
	req = muxSetVars(req, map[string]string{"matchId": "3"})
	rr := httptest.NewRecorder()
	GetMatchParticipants(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 InternalServerError, got %d", rr.Code)
	}
}

func TestSaveMatchResults_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE match_participants").
		WithArgs(10, true, 2, 5).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE matches SET completed_at = NOW\\(\\) WHERE match_id = \\$1").
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	body := `[{"participant_id":5,"score":10,"is_winner":true}]`
	req := httptest.NewRequest(http.MethodPut, "/api/matches/2/results", bytes.NewReader([]byte(body)))
	req = muxSetVars(req, map[string]string{"matchId": "2"})
	rr := httptest.NewRecorder()
	SaveMatchResults(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestSaveMatchResults_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPut, "/api/matches/abc/results", nil)
	req = muxSetVars(req, map[string]string{"matchId": "abc"})
	rr := httptest.NewRecorder()
	SaveMatchResults(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestSaveMatchResults_BadJSON(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPut, "/api/matches/2/results", bytes.NewReader([]byte("bad json")))
	req = muxSetVars(req, map[string]string{"matchId": "2"})
	rr := httptest.NewRecorder()
	SaveMatchResults(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestSaveMatchResults_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE match_participants").
		WithArgs(10, true, 2, 5).
		WillReturnError(errors.New("db fail"))
	mock.ExpectRollback()
	body := `[{"participant_id":5,"score":10,"is_winner":true}]`
	req := httptest.NewRequest(http.MethodPut, "/api/matches/2/results", bytes.NewReader([]byte(body)))
	req = muxSetVars(req, map[string]string{"matchId": "2"})
	rr := httptest.NewRecorder()
	SaveMatchResults(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 InternalServerError, got %d", rr.Code)
	}
}

func TestGenerateNextRound_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/stages/abc/rounds", nil)
	req = muxSetVars(req, map[string]string{"stageId": "abc"})
	rr := httptest.NewRecorder()
	GenerateNextRound(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestCanGenerateNextRound_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodGet, "/api/stages/abc/can-generate-next-round", nil)
	req = muxSetVars(req, map[string]string{"stageId": "abc"})
	rr := httptest.NewRecorder()
	CanGenerateNextRound(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestCanGenerateNextRound_StageNotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT competition_id FROM competition_stages").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)
	req := httptest.NewRequest(http.MethodGet, "/api/stages/1/can-generate-next-round", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	CanGenerateNextRound(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestCanGenerateNextRound_CompetitionNotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT competition_id FROM competition_stages").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"competition_id"}).AddRow(2))
	mock.ExpectQuery("SELECT status FROM competitions").
		WithArgs(2).
		WillReturnError(sql.ErrNoRows)
	req := httptest.NewRequest(http.MethodGet, "/api/stages/1/can-generate-next-round", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	CanGenerateNextRound(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestCanGenerateNextRound_NotOngoing(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT competition_id FROM competition_stages").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"competition_id"}).AddRow(2))
	mock.ExpectQuery("SELECT status FROM competitions").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(1)) // Not 2
	req := httptest.NewRequest(http.MethodGet, "/api/stages/1/can-generate-next-round", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	CanGenerateNextRound(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestCanGenerateNextRound_NoRounds(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT competition_id FROM competition_stages").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"competition_id"}).AddRow(2))
	mock.ExpectQuery("SELECT status FROM competitions").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(2))
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(round_number\\), 0\\) FROM rounds WHERE stage_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(0))
	req := httptest.NewRequest(http.MethodGet, "/api/stages/1/can-generate-next-round", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	CanGenerateNextRound(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestCanGenerateNextRound_UnfinishedMatches(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT competition_id FROM competition_stages").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"competition_id"}).AddRow(2))
	mock.ExpectQuery("SELECT status FROM competitions").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(2))
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(round_number\\), 0\\) FROM rounds WHERE stage_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(1))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM matches").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	req := httptest.NewRequest(http.MethodGet, "/api/stages/1/can-generate-next-round", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	CanGenerateNextRound(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestAdvanceAfterRoundRobin_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/stages/abc/advance", nil)
	req = muxSetVars(req, map[string]string{"stageId": "abc"})
	rr := httptest.NewRecorder()
	AdvanceAfterRoundRobin(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestAdvanceAfterRoundRobin_UnfinishedMatches(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM matches m").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	req := httptest.NewRequest(http.MethodPost, "/api/stages/1/advance", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	AdvanceAfterRoundRobin(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestAdvanceAfterRoundRobin_NoNextStage_FinishCompetition(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM matches m").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT stage_id, participants_at_start FROM competition_stages").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("UPDATE competitions SET status = 3").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	req := httptest.NewRequest(http.MethodPost, "/api/stages/1/advance", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	AdvanceAfterRoundRobin(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestAdvanceAfterRoundRobin_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM matches m").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT stage_id, participants_at_start FROM competition_stages").
		WithArgs(1).
		WillReturnError(errors.New("db fail"))
	req := httptest.NewRequest(http.MethodPost, "/api/stages/1/advance", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	AdvanceAfterRoundRobin(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 InternalServerError, got %d", rr.Code)
	}
}

func TestGenerateNextRound_DBErrorOnLastRound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(round_number\\), 0\\) FROM rounds WHERE stage_id = \\$1").
		WithArgs(1).
		WillReturnError(errors.New("db fail"))
	req := httptest.NewRequest(http.MethodPost, "/api/stages/1/rounds", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	GenerateNextRound(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 InternalServerError, got %d", rr.Code)
	}
}

func TestGenerateNextRound_UnfinishedMatches(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(round_number\\), 0\\) FROM rounds WHERE stage_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(2))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM matches").
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	req := httptest.NewRequest(http.MethodPost, "/api/stages/1/rounds", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	GenerateNextRound(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestGenerateNextRound_DBErrorOnFormat(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(round_number\\), 0\\) FROM rounds WHERE stage_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(0))
	mock.ExpectQuery("SELECT tourney_format_id FROM competition_stages WHERE stage_id=\\$1").
		WithArgs(1).
		WillReturnError(errors.New("db fail"))
	req := httptest.NewRequest(http.MethodPost, "/api/stages/1/rounds", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	GenerateNextRound(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 InternalServerError, got %d", rr.Code)
	}
}

func TestGenerateNextRound_UnsupportedFormat(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(round_number\\), 0\\) FROM rounds WHERE stage_id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(0))
	mock.ExpectQuery("SELECT tourney_format_id FROM competition_stages WHERE stage_id=\\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"tourney_format_id"}).AddRow(99))
	req := httptest.NewRequest(http.MethodPost, "/api/stages/1/rounds", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	GenerateNextRound(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}
