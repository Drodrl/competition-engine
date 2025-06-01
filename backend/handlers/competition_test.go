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

func ptr(s string) *string { return &s }
func ptrInt(i int) *int    { return &i }

func TestCreateDraftCompetition_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	competition := models.Competition{
		CompetitionName: "Test",
		SportID:         1,
		StartDate:       ptr("2024-01-01"),
		EndDate:         ptr("2024-01-02"),
		OrganizerID:     1,
		MaxParticipants: ptrInt(16),
		FlagTeams:       false,
	}
	mock.ExpectQuery("INSERT INTO competitions").
		WithArgs(competition.CompetitionName, competition.SportID, competition.StartDate, competition.EndDate, competition.OrganizerID, sqlmock.AnyArg(), competition.MaxParticipants, competition.FlagTeams).
		WillReturnRows(sqlmock.NewRows([]string{"competition_id"}).AddRow(42))

	body, _ := json.Marshal(competition)
	req := httptest.NewRequest(http.MethodPost, "/api/competitions/draft", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	CreateDraftCompetition(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var resp map[string]int
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["competition_id"] != 42 {
		t.Errorf("expected competition_id 42, got %d", resp["competition_id"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateDraftCompetition_BadJSON(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/competitions/draft", bytes.NewReader([]byte("bad json")))
	rr := httptest.NewRecorder()
	CreateDraftCompetition(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestDeleteCompetition_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec("DELETE FROM competitions").WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM competition_stages").WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM competition_participants").WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodDelete, "/api/competitions/1", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	DeleteCompetition(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 NoContent, got %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDeleteCompetition_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodDelete, "/api/competitions/abc", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "abc"})
	rr := httptest.NewRecorder()
	DeleteCompetition(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestGetCompetitionsByOrganizer_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT competition_id, competition_name").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{
			"competition_id", "competition_name", "sport_id", "start_date", "end_date", "max_participants", "organizer_id", "status", "date_created", "date_updated", "flag_teams",
		}).AddRow(1, "Comp1", 2, "2024-01-01", "2024-01-02", 16, 123, 0, "2024-01-01", "2024-01-01", false))

	req := httptest.NewRequest(http.MethodGet, "/api/competitions/organizer/123", nil)
	req = muxSetVars(req, map[string]string{"organizerId": "123"})
	rr := httptest.NewRecorder()
	GetCompetitionsByOrganizer(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var comps []models.Competition
	if err := json.NewDecoder(rr.Body).Decode(&comps); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(comps) != 1 || comps[0].CompetitionId != 1 {
		t.Errorf("unexpected competitions: %+v", comps)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetCompetitionsByOrganizer_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodGet, "/api/competitions/organizer/abc", nil)
	req = muxSetVars(req, map[string]string{"organizerId": "abc"})
	rr := httptest.NewRecorder()
	GetCompetitionsByOrganizer(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestGetCompetitionByID_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodGet, "/api/competitions/abc", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "abc"})
	rr := httptest.NewRecorder()
	GetCompetitionByID(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestUpdateCompetition_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPut, "/api/competitions/abc", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "abc"})
	rr := httptest.NewRecorder()
	UpdateCompetition(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestUpdateCompetition_BadJSON(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPut, "/api/competitions/1", bytes.NewReader([]byte("bad json")))
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	UpdateCompetition(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestGetAllCompetitions_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT competition_id, competition_name").
		WillReturnRows(sqlmock.NewRows([]string{
			"competition_id", "competition_name", "sport_id", "start_date", "end_date", "max_participants", "organizer_id", "status", "date_created", "date_updated", "flag_teams",
		}).AddRow(1, "Comp1", 2, "2024-01-01", "2024-01-02", 16, 123, 0, "2024-01-01", "2024-01-01", false))

	req := httptest.NewRequest(http.MethodGet, "/api/competitions", nil)
	rr := httptest.NewRecorder()
	GetAllCompetitions(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var comps []models.Competition
	if err := json.NewDecoder(rr.Body).Decode(&comps); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(comps) != 1 || comps[0].CompetitionId != 1 {
		t.Errorf("unexpected competitions: %+v", comps)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetCompetitionsByFlagTeams_BadFlag(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodGet, "/api/competitions/flag_teams/notabool", nil)
	req = muxSetVars(req, map[string]string{"flagTeams": "notabool"})
	rr := httptest.NewRecorder()
	GetCompetitionsByFlagTeams(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestGetCompetitionsByFlagTeams_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT competition_id, competition_name").
		WithArgs(true).
		WillReturnRows(sqlmock.NewRows([]string{
			"competition_id", "competition_name", "sport_id", "start_date", "end_date", "max_participants", "organizer_id", "status", "date_created", "date_updated", "flag_teams",
		}).AddRow(1, "Comp1", 2, "2024-01-01", "2024-01-02", 16, 123, 0, "2024-01-01", "2024-01-01", true))

	req := httptest.NewRequest(http.MethodGet, "/api/competitions/flag_teams/true", nil)
	req = muxSetVars(req, map[string]string{"flagTeams": "true"})
	rr := httptest.NewRecorder()
	GetCompetitionsByFlagTeams(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var comps []models.Competition
	if err := json.NewDecoder(rr.Body).Decode(&comps); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(comps) != 1 || comps[0].CompetitionId != 1 {
		t.Errorf("unexpected competitions: %+v", comps)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetCompetitionByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT c.competition_id, c.competition_name").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"competition_id", "competition_name", "sport_id", "start_date", "end_date", "max_participants", "organizer_id", "status", "date_created", "date_updated", "flag_teams", "sport_name",
		}).AddRow(1, "Comp1", 2, "2024-01-01", "2024-01-02", 16, 123, 1, "2024-01-01", "2024-01-01", false, "Soccer"))

	req := httptest.NewRequest(http.MethodGet, "/api/competitions/1", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	GetCompetitionByID(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if int(resp["competition_id"].(float64)) != 1 {
		t.Errorf("unexpected competition_id: %+v", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdateCompetition_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec("UPDATE competitions").
		WithArgs("New Name", ptr("2024-01-01"), ptr("2024-01-02"), ptrInt(10), true, sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payload := map[string]interface{}{
		"competition_name": "New Name",
		"start_date":       "2024-01-01",
		"end_date":         "2024-01-02",
		"max_participants": 10,
		"flag_teams":       true,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/api/competitions/1", bytes.NewReader(body))
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	UpdateCompetition(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestChangeCompetitionStatus_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPatch, "/api/competitions/abc/status", bytes.NewReader([]byte(`{"status":1}`)))
	req = muxSetVars(req, map[string]string{"competitionId": "abc"})
	rr := httptest.NewRecorder()
	ChangeCompetitionStatus(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestGetStagesByCompetitionID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT stage_id, stage_name").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"stage_id", "stage_name", "stage_order", "tourney_format_id", "participants_at_start", "participants_at_end",
		}).AddRow(1, "Stage 1", 1, 1, 8, 4))

	req := httptest.NewRequest(http.MethodGet, "/api/competitions/1/stages", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	GetStagesByCompetitionID(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var stages []models.StageDTO
	if err := json.NewDecoder(rr.Body).Decode(&stages); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(stages) != 1 || stages[0].StageID != 1 {
		t.Errorf("unexpected stages: %+v", stages)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAddStageToCompetition_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/competitions/abc/stages", bytes.NewReader([]byte("{}")))
	req = muxSetVars(req, map[string]string{"competitionId": "abc"})
	rr := httptest.NewRecorder()
	AddStageToCompetition(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestUpdateStage_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPut, "/api/competitions/abc/stages/1", bytes.NewReader([]byte("{}")))
	req = muxSetVars(req, map[string]string{"competitionId": "abc", "stageId": "1"})
	rr := httptest.NewRecorder()
	UpdateStage(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestDeleteStage_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodDelete, "/api/competitions/1/stages/abc", nil)
	req = muxSetVars(req, map[string]string{"stageId": "abc"})
	rr := httptest.NewRecorder()
	DeleteStage(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestGetParticipantsByCompetitionID_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodGet, "/api/competitions/abc/participants", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "abc"})
	rr := httptest.NewRecorder()
	GetParticipantsByCompetitionID(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestFinishCompetition_BadID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/competitions/abc/finish", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "abc"})
	rr := httptest.NewRecorder()
	FinishCompetition(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestAddStageToCompetition_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT stage_id, stage_name").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"stage_id", "stage_name", "stage_order", "tourney_format_id", "participants_at_start", "participants_at_end",
		}))

	mock.ExpectQuery("SELECT max_participants FROM competitions").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"max_participants"}).AddRow(8))

	mock.ExpectQuery("SELECT minimum_participants FROM tournament_formats").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"minimum_participants"}).AddRow(2))

	mock.ExpectExec("INSERT INTO competition_stages").
		WithArgs(1, 1, "Stage 1", 1, 8, 4).
		WillReturnResult(sqlmock.NewResult(1, 1))

	stage := models.StageDTO{
		StageName:           "Stage 1",
		StageOrder:          1,
		TourneyFormatID:     1,
		ParticipantsAtStart: 8,
		ParticipantsAtEnd:   4,
	}
	body, _ := json.Marshal(stage)
	req := httptest.NewRequest(http.MethodPost, "/api/competitions/1/stages", bytes.NewReader(body))
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	AddStageToCompetition(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAddStageToCompetition_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT stage_id, stage_name").
		WithArgs(1).
		WillReturnError(errors.New("db fail"))

	stage := models.StageDTO{}
	body, _ := json.Marshal(stage)
	req := httptest.NewRequest(http.MethodPost, "/api/competitions/1/stages", bytes.NewReader(body))
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	AddStageToCompetition(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 InternalServerError, got %d", rr.Code)
	}
}

func TestAddStageToCompetition_BusinessRuleFail(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT stage_id, stage_name").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"stage_id", "stage_name", "stage_order", "tourney_format_id", "participants_at_start", "participants_at_end",
		}).AddRow(1, "Stage 1", 1, 1, 8, 4))

	mock.ExpectQuery("SELECT max_participants FROM competitions").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"max_participants"}).AddRow(8))

	mock.ExpectQuery("SELECT minimum_participants FROM tournament_formats").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"minimum_participants"}).AddRow(2))

	stage := models.StageDTO{
		StageName:           "Stage 2",
		StageOrder:          2,
		TourneyFormatID:     2,
		ParticipantsAtStart: 7,
		ParticipantsAtEnd:   1,
	}
	body, _ := json.Marshal(stage)
	req := httptest.NewRequest(http.MethodPost, "/api/competitions/1/stages", bytes.NewReader(body))
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	AddStageToCompetition(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestUpdateStage_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT stage_id, stage_name").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"stage_id", "stage_name", "stage_order", "tourney_format_id", "participants_at_start", "participants_at_end",
		}).AddRow(1, "Stage 1", 1, 1, 8, 4))

	mock.ExpectQuery("SELECT max_participants FROM competitions").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"max_participants"}).AddRow(8))

	mock.ExpectQuery("SELECT minimum_participants FROM tournament_formats").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"minimum_participants"}).AddRow(2))

	mock.ExpectExec("UPDATE competition_stages").
		WithArgs("Stage 1", 1, 1, 8, 4, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	stage := models.StageDTO{
		StageName:           "Stage 1",
		StageOrder:          1,
		TourneyFormatID:     1,
		ParticipantsAtStart: 8,
		ParticipantsAtEnd:   4,
	}
	body, _ := json.Marshal(stage)
	req := httptest.NewRequest(http.MethodPut, "/api/competitions/1/stages/1", bytes.NewReader(body))
	req = muxSetVars(req, map[string]string{"competitionId": "1", "stageId": "1"})
	rr := httptest.NewRecorder()
	UpdateStage(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDeleteStage_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec("DELETE FROM competition_stages").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodDelete, "/api/competitions/1/stages/1", nil)
	req = muxSetVars(req, map[string]string{"stageId": "1"})
	rr := httptest.NewRecorder()
	DeleteStage(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestChangeCompetitionStatus_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT status, max_participants, sport_id FROM competitions").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"status", "max_participants", "sport_id"}).AddRow(0, 8, 1))

	mock.ExpectQuery("SELECT stage_id, stage_name").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"stage_id", "stage_name", "stage_order", "tourney_format_id", "participants_at_start", "participants_at_end",
		}).AddRow(1, "Stage 1", 1, 1, 8, 4))

	mock.ExpectQuery("SELECT tourney_format_id, min_participants FROM tournament_formats").
		WillReturnRows(sqlmock.NewRows([]string{"tourney_format_id", "min_participants"}).AddRow(1, 2))

	mock.ExpectExec("UPDATE competitions").
		WithArgs(1, sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payload := map[string]interface{}{"status": 1}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPatch, "/api/competitions/1/status", bytes.NewReader(body))
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	ChangeCompetitionStatus(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestFinishCompetition_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT stage_id FROM competition_stages").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"stage_id"}).AddRow(2))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM matches m").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT round_id FROM rounds").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"round_id"}).AddRow(3))

	mock.ExpectQuery("SELECT mp.user_id, mp.team_id").
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "team_id"}).AddRow(5, 7))

	mock.ExpectQuery("SELECT name_user FROM users").
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"name_user"}).AddRow("Winner User"))

	mock.ExpectQuery("SELECT team_name FROM teams").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"team_name"}).AddRow("Winner Team"))

	mock.ExpectExec("UPDATE competitions SET status = 3").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/api/competitions/1/finish", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	FinishCompetition(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !resp["finished"].(bool) {
		t.Errorf("expected finished true, got %+v", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetStagesByCompetitionID_Empty(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT stage_id, stage_name").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"stage_id", "stage_name", "stage_order", "tourney_format_id", "participants_at_start", "participants_at_end",
		}))

	req := httptest.NewRequest(http.MethodGet, "/api/competitions/1/stages", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	GetStagesByCompetitionID(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var stages []models.StageDTO
	if err := json.NewDecoder(rr.Body).Decode(&stages); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(stages) != 0 {
		t.Errorf("expected 0 stages, got %+v", stages)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetParticipantsByCompetitionID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT u.id_user, u.name_user, u.lname1_user").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id_user", "name_user", "lname1_user"}).
			AddRow(10, "Alice", "Smith").
			AddRow(11, "Bob", "Jones"))

	mock.ExpectQuery("SELECT t.team_id, t.team_name, NULL").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"team_id", "team_name", "lname1_user"}).
			AddRow(20, "TeamX", nil))

	req := httptest.NewRequest(http.MethodGet, "/api/competitions/1/participants", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	GetParticipantsByCompetitionID(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var participants []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&participants); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(participants) == 0 {
		t.Errorf("expected participants, got none")
	}
}

func TestGetParticipantsByCompetitionID_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT u.id_user, u.name_user, u.lname1_user").
		WithArgs(1).
		WillReturnError(errors.New("db fail"))

	req := httptest.NewRequest(http.MethodGet, "/api/competitions/1/participants", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	GetParticipantsByCompetitionID(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 InternalServerError, got %d", rr.Code)
	}
}

func TestFinishCompetition_NoStages(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT stage_id FROM competition_stages").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodPost, "/api/competitions/1/finish", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	FinishCompetition(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestFinishCompetition_UnfinishedMatches(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT stage_id FROM competition_stages").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"stage_id"}).AddRow(2))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM matches m").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1)) // unfinished > 0

	req := httptest.NewRequest(http.MethodPost, "/api/competitions/1/finish", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	FinishCompetition(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestFinishCompetition_NoWinner(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT stage_id FROM competition_stages").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"stage_id"}).AddRow(2))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM matches m").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT round_id FROM rounds").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"round_id"}).AddRow(3))

	mock.ExpectQuery("SELECT mp.user_id, mp.team_id").
		WithArgs(3).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodPost, "/api/competitions/1/finish", nil)
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	FinishCompetition(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestChangeCompetitionStatus_CloseSignup_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT status, max_participants, sport_id FROM competitions").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"status", "max_participants", "sport_id"}).AddRow(1, 8, 1))

	mock.ExpectQuery("SELECT max_participants FROM competitions").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"max_participants"}).AddRow(8))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM competition_participants").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))

	mock.ExpectQuery("SELECT stage_id FROM competition_stages").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"stage_id"}).AddRow(2))

	mock.ExpectExec("INSERT INTO stage_participants \\(stage_id, user_id\\)").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO stage_participants \\(stage_id, team_id\\)").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE competitions SET status = .*date_updated = .*WHERE competition_id = .*").
		WithArgs(2, sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payload := map[string]interface{}{"status": 2}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPatch, "/api/competitions/1/status", bytes.NewReader(body))
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	ChangeCompetitionStatus(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestChangeCompetitionStatus_CloseSignup_NotEnoughParticipants(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT status, max_participants, sport_id FROM competitions").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"status", "max_participants", "sport_id"}).AddRow(1, 8, 1))

	mock.ExpectQuery("SELECT max_participants FROM competitions").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"max_participants"}).AddRow(8))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM competition_participants").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	payload := map[string]interface{}{"status": 2}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPatch, "/api/competitions/1/status", bytes.NewReader(body))
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	ChangeCompetitionStatus(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}

func TestChangeCompetitionStatus_CloseSignup_NoStages(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT status, max_participants, sport_id FROM competitions").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"status", "max_participants", "sport_id"}).AddRow(1, 8, 1))

	mock.ExpectQuery("SELECT max_participants FROM competitions").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"max_participants"}).AddRow(8))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM competition_participants").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))

	mock.ExpectQuery("SELECT stage_id FROM competition_stages").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	payload := map[string]interface{}{"status": 2}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPatch, "/api/competitions/1/status", bytes.NewReader(body))
	req = muxSetVars(req, map[string]string{"competitionId": "1"})
	rr := httptest.NewRecorder()
	ChangeCompetitionStatus(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", rr.Code)
	}
}
