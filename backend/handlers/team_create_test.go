package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestTeamCreateSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Mock the database queries
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM teams WHERE team_name=\\$1\\)").
		WithArgs("Test Team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery("INSERT INTO teams \\(team_name, date_created, date_updated\\) VALUES \\(\\$1, \\$2, \\$3\\) RETURNING team_id").
		WithArgs("Test Team", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(1))

	mock.ExpectExec("INSERT INTO user_teams \\(team_id, user_id, team_position, date_created, date_updated\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\)").
		WithArgs(1, 123, "Team Leader", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	handler := NewTeamCreateHandler(db)

	payload := CreateTeamRequest{
		TeamName: "Test Team",
		UserIDs:  []int{123},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/team_create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rec.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestTeamCreateNameExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Mock team name existance
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM teams WHERE team_name=\\$1\\)").
		WithArgs("Existing Team").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	handler := NewTeamCreateHandler(db)

	payload := CreateTeamRequest{
		TeamName: "Existing Team",
		UserIDs:  []int{123},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/team_create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
	}

	expected := "Team name already exists"
	if rec.Body.String() != expected+"\n" {
		t.Errorf("Expected response body %q, got %q", expected, rec.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestTeamCreateInvalidPayload(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	handler := NewTeamCreateHandler(db)

	body := []byte(`{invalid json}`)

	req := httptest.NewRequest(http.MethodPost, "/team_create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
	}

	expected := "Invalid request payload"
	if rec.Body.String() != expected+"\n" {
		t.Errorf("Expected response body %q, got %q", expected, rec.Body.String())
	}
}
