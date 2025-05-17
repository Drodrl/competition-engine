package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestLoginHandlerSuccess(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	creds := Credentials{Email: "u@e.com", Password: "pw"}
	body, _ := json.Marshal(creds)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewLoginHandler(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK; got %d", rr.Code)
	}

	var resp LoginResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Message != "Login successful" {
		t.Errorf("expected message 'Login successful'; got %q", resp.Message)
	}
}

func TestLoginHandlerMethodNotAllowed(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rr := httptest.NewRecorder()

	NewLoginHandler(db).ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 MethodNotAllowed; got %d", rr.Code)
	}
}

func TestLoginHandlerBadJSON(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(`not-json`)))
	rr := httptest.NewRecorder()

	NewLoginHandler(db).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest; got %d", rr.Code)
	}
}
