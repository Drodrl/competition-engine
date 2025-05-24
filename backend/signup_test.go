package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// User Signup Tests

func TestUserSignupSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	userID := 45
	competitionID := 12
	reqBody := UserSignupRequest{
		CompetitionID: competitionID,
		UserID:        &userID,
	}
	body, _ := json.Marshal(reqBody)

	// Mock user existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id=\$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Mock competition existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competitions WHERE competition_id=\$1\)`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectExec(`INSERT INTO competition_participants`).
		WithArgs(competitionID, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/user_signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewUserSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created; got %d", rr.Code)
	}
}

func TestSignupUserDoesNotExist(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	userID := 45
	competitionID := 12
	reqBody := UserSignupRequest{
		CompetitionID: competitionID,
		UserID:        &userID,
	}
	body, _ := json.Marshal(reqBody)

	// Simulate user does not exist
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id_user=\$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	req := httptest.NewRequest(http.MethodPost, "/user_signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewUserSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest; got %d", rr.Code)
	}
}

func TestUserSignupCompetitionDoesNotExist(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	userID := 45
	competitionID := 12
	reqBody := UserSignupRequest{
		CompetitionID: competitionID,
		UserID:        &userID,
	}
	body, _ := json.Marshal(reqBody)

	// Mock user existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id_user=\$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Simulate competition does not exist
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competitions WHERE competition_id=\$1\)`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	req := httptest.NewRequest(http.MethodPost, "/user_signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewUserSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest; got %d", rr.Code)
	}
}

func TestUserSignupMethodNotAllowed(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/user_signup", nil)
	rr := httptest.NewRecorder()

	handler := NewUserSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 MethodNotAllowed; got %d", rr.Code)
	}
}

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
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE id=\$1\)`).
		WithArgs(teamID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Mock competition existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competitions WHERE competition_id=\$1\)`).
		WithArgs(competitionID).
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
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE id=\$1\)`).
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
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE id_team=\$1\)`).
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
