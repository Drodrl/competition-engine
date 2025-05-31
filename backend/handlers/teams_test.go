package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	return db, mock
}

func TestNewTeamsHandler(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	handler := NewTeamsHandler(db)

	userID := "1"
	mock.ExpectQuery(`SELECT t.team_id, t.team_name FROM user_teams ut JOIN teams t ON ut.team_id = t.team_id WHERE ut.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"team_id", "team_name"}).
			AddRow(1, "Team A").
			AddRow(2, "Team B"))

	req := httptest.NewRequest(http.MethodGet, "/api/teams?user_id=1", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var teams []struct {
		TeamID   int    `json:"team_id"`
		TeamName string `json:"team_name"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&teams); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(teams) != 2 || teams[0].TeamName != "Team A" || teams[1].TeamName != "Team B" {
		t.Errorf("Unexpected response: %+v", teams)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestRemoveParticipantsHandler(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	handler := RemoveParticipantsHandler(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM user_teams WHERE team_id = \\$1 AND user_id = \\$2").
		WithArgs(1, 10).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE teams SET date_updated = NOW\\(\\) WHERE team_id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	payload := map[string]interface{}{
		"team_id":  1,
		"user_ids": []int{10},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/remove-participants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestAddParticipantsHandler(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	handler := AddParticipantsHandler(db)

	mock.ExpectBegin()

	// Mock the query to check if the user exists in the users table
	mock.ExpectQuery("SELECT EXISTS \\(SELECT 1 FROM users WHERE id_user = \\$1\\)").
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock the query to check if the user is already in the team
	mock.ExpectQuery("SELECT EXISTS \\(SELECT 1 FROM user_teams WHERE team_id = \\$1 AND user_id = \\$2\\)").
		WithArgs(1, 10).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock the query to insert the user into the team
	mock.ExpectExec("INSERT INTO user_teams \\(user_id, team_id, date_updated\\) VALUES \\(\\$1, \\$2, NOW\\(\\)\\)").
		WithArgs(10, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock the query to update the team's date_updated field
	mock.ExpectExec("UPDATE teams SET date_updated = NOW\\(\\) WHERE team_id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	payload := map[string]interface{}{
		"team_id":  1,
		"user_ids": []int{10},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/add-participants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}
