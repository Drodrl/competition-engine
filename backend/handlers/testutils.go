package handlers

import (
	"net/http"
	"testing"

	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	SetDB(db)
	return db, mock
}

func muxSetVars(r *http.Request, vars map[string]string) *http.Request {
	return mux.SetURLVars(r, vars)
}
