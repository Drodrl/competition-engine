package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type SignUpRequest struct {
	StageID int  `json:"stage_id"`
	UserID  *int `json:"user_id,omitempty"`
	TeamID  *int `json:"team_id,omitempty"`
}

func TestSignupUserSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	userID := 45
	stageID := 12
	reqBody := SignUpRequest{
		StageID: stageID,
		UserID:  &userID,
	}
	body, _ := json.Marshal(reqBody)

	// Mock user existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id=\$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectExec(`INSERT INTO stage_participants`).
		WithArgs(stageID, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created; got %d", rr.Code)
	}
}

func TestSignupTeamSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	teamID := 99
	stageID := 12
	reqBody := SignUpRequest{
		StageID: stageID,
		TeamID:  &teamID,
	}
	body, _ := json.Marshal(reqBody)

	// Mock team existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE id=\$1\)`).
		WithArgs(teamID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectExec(`INSERT INTO stage_participants`).
		WithArgs(stageID, teamID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created; got %d", rr.Code)
	}
}

func TestSignupMissingUserAndTeam(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	reqBody := SignUpRequest{
		StageID: 10,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler := NewSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest; got %d", rr.Code)
	}
}

func TestSignupBothUserAndTeamProvided(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	uid := 1
	tid := 2
	reqBody := SignUpRequest{
		StageID: 5,
		UserID:  &uid,
		TeamID:  &tid,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler := NewSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest; got %d", rr.Code)
	}
}

func TestSignupMethodNotAllowed(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/signup", nil)
	rr := httptest.NewRecorder()

	handler := NewSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 MethodNotAllowed; got %d", rr.Code)
	}
}

func TestSignupUserDoesNotExist(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	userID := 45
	stageID := 12
	reqBody := SignUpRequest{
		StageID: stageID,
		UserID:  &userID,
	}
	body, _ := json.Marshal(reqBody)

	// Simulate that the user does not exist
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id=\$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest; got %d", rr.Code)
	}
}

func TestSignupTeamDoesNotExist(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	teamID := 99
	stageID := 12
	reqBody := SignUpRequest{
		StageID: stageID,
		TeamID:  &teamID,
	}
	body, _ := json.Marshal(reqBody)

	// Simulate that the team does not exist
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE id=\$1\)`).
		WithArgs(teamID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest; got %d", rr.Code)
	}
}
