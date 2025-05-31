package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// Team Signup Tests

func TestTeamSignupSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	teamID := 99
	competitionID := 12
	reqBody := TeamSignupRequest{
		CompetitionID: competitionID,
		TeamID:        &teamID,
	}
	body, _ := json.Marshal(reqBody)

	// Mock team existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE team_id=\$1\)`).
		WithArgs(teamID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock competition existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competitions WHERE competition_id=\$1\)`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock competition status check
	mock.ExpectQuery(`SELECT status FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(1))

	// Mock check for existing signup
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competition_participants WHERE competition_id=\$1 AND team_id=\$2\)`).
		WithArgs(competitionID, teamID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock max participants
	mock.ExpectQuery(`SELECT max_participants FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"max_participants"}).AddRow(5))

	// Mock current participants count
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM competition_participants WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock user is a team leader
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_teams WHERE team_id=\$1 AND team_position='Team Leader'\)`).
		WithArgs(teamID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectExec(`INSERT INTO competition_participants`).
		WithArgs(competitionID, teamID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/team_signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewTeamSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created; got %d", rr.Code)
	}
}

func TestSignupTeamDoesNotExist(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	teamID := 99
	competitionID := 12
	reqBody := TeamSignupRequest{
		CompetitionID: competitionID,
		TeamID:        &teamID,
	}
	body, _ := json.Marshal(reqBody)

	// Simulate that the team does not exist
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE team_id=\$1\)`).
		WithArgs(teamID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	req := httptest.NewRequest(http.MethodPost, "/team_signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewTeamSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest; got %d", rr.Code)
	}
}

func TestTeamSignupCompetitionDoesNotExist(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	teamID := 99
	competitionID := 12
	reqBody := TeamSignupRequest{
		CompetitionID: competitionID,
		TeamID:        &teamID,
	}
	body, _ := json.Marshal(reqBody)

	// Mock team existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE team_id=\$1\)`).
		WithArgs(teamID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Simulate competition does not exist
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competitions WHERE competition_id=\$1\)`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	req := httptest.NewRequest(http.MethodPost, "/team_signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewTeamSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest; got %d", rr.Code)
	}
}

func TestTeamSignupCompetitionFull(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	teamID := 45
	competitionID := 12
	reqBody := TeamSignupRequest{
		CompetitionID: competitionID,
		TeamID:        &teamID,
	}
	body, _ := json.Marshal(reqBody)

	// Mock team existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE team_id=\$1\)`).
		WithArgs(teamID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock competition existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competitions WHERE competition_id=\$1\)`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock competition status check
	mock.ExpectQuery(`SELECT status FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(1))

	// Mock team is not already signed up
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competition_participants WHERE competition_id=\$1 AND team_id=\$2\)`).
		WithArgs(competitionID, teamID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock max participants
	mock.ExpectQuery(`SELECT max_participants FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"max_participants"}).AddRow(2))

	// Mock current participants count (full)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM competition_participants WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	req := httptest.NewRequest(http.MethodPost, "/team_signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewTeamSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest; got %d", rr.Code)
	}
}

func TestTeamSignupMethodNotAllowed(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/team_signup/team", nil)
	rr := httptest.NewRecorder()

	handler := NewTeamSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 MethodNotAllowed; got %d", rr.Code)
	}
}
