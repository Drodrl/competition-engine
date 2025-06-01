package handlers

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
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id_user=\$1\)`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock competition existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competitions WHERE competition_id=\$1\)`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock competition status check
	mock.ExpectQuery(`SELECT status FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(1))

	// Mock it is not a team competition
	mock.ExpectQuery(`SELECT flag_teams FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"flag_teams"}).AddRow(false))

	// Mock user is not already signed up
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competition_participants WHERE competition_id=\$1 AND user_id=\$2\)`).
		WithArgs(competitionID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock max participants
	mock.ExpectQuery(`SELECT max_participants FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"max_participants"}).AddRow(100))

	// Mock current participants count
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM competition_participants WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

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

func TestUserSignupCompetitionFull(t *testing.T) {
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

	// Mock competition existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competitions WHERE competition_id=\$1\)`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock competition status check
	mock.ExpectQuery(`SELECT status FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(1))

	// Mock single player competition
	mock.ExpectQuery(`SELECT flag_teams FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"flag_teams"}).AddRow(false))

	// Mock user is not already signed up
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competition_participants WHERE competition_id=\$1 AND user_id=\$2\)`).
		WithArgs(competitionID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock max participants
	mock.ExpectQuery(`SELECT max_participants FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"max_participants"}).AddRow(2))

	// Mock current participants count (full)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM competition_participants WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	req := httptest.NewRequest(http.MethodPost, "/user_signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewUserSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest; got %d", rr.Code)
	}
}

func TestUserSignupCompetitionNotOpen(t *testing.T) {
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

	// Mock competition existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competitions WHERE competition_id=\$1\)`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Simulate competition is not open
	mock.ExpectQuery(`SELECT status FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(0))

	req := httptest.NewRequest(http.MethodPost, "/user_signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewUserSignupHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest; got %d", rr.Code)
	}
}

func TestUserSignupTeamCompetition(t *testing.T) {
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

	// Mock competition existence
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM competitions WHERE competition_id=\$1\)`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock competition status check
	mock.ExpectQuery(`SELECT status FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(1))

	// Mock it is a team competition
	mock.ExpectQuery(`SELECT flag_teams FROM competitions WHERE competition_id=\$1`).
		WithArgs(competitionID).
		WillReturnRows(sqlmock.NewRows([]string{"flag_teams"}).AddRow(true))

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
