package controllers

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return db, mock
}

// --- GenerateRoundSingleElim ---

func TestGenerateRoundSingleElim_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	stageID := 1

	mock.ExpectQuery(`SELECT COALESCE\(MAX\(round_number\), 0\) \+ 1 FROM rounds WHERE stage_id = \$1`).
		WithArgs(stageID).
		WillReturnRows(sqlmock.NewRows([]string{"next_round"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"user_id", "team_id"}).
		AddRow(1, nil).
		AddRow(2, nil)
	mock.ExpectQuery(`SELECT user_id, team_id FROM stage_participants WHERE stage_id=\$1`).
		WithArgs(stageID).
		WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO rounds \(stage_id, round_number\) VALUES \(\$1,\$2\) RETURNING round_id`).
		WithArgs(stageID, 1).WillReturnRows(sqlmock.NewRows([]string{"round_id"}).AddRow(10))

	mock.ExpectQuery(`INSERT INTO matches \(round_id, scheduled_at\) VALUES \(\$1, NOW\(\)\) RETURNING match_id`).
		WithArgs(10).WillReturnRows(sqlmock.NewRows([]string{"match_id"}).AddRow(100))

	mock.ExpectExec(`INSERT INTO match_participants`).WithArgs(
		100, 1, nil, 2, nil,
	).WillReturnResult(sqlmock.NewResult(1, 2))

	mock.ExpectCommit()

	err := GenerateRoundSingleElim(db, stageID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateRoundSingleElim_OddParticipants(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	stageID := 1
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(round_number\), 0\) \+ 1 FROM rounds WHERE stage_id = \$1`).
		WithArgs(stageID).
		WillReturnRows(sqlmock.NewRows([]string{"next_round"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"user_id", "team_id"}).
		AddRow(1, nil).
		AddRow(2, nil).
		AddRow(3, nil)
	mock.ExpectQuery(`SELECT user_id, team_id FROM stage_participants WHERE stage_id=\$1`).
		WithArgs(stageID).
		WillReturnRows(rows)

	err := GenerateRoundSingleElim(db, stageID)
	if err == nil || err.Error() != "expected even participants, got 3" {
		t.Errorf("expected error for odd participants, got: %v", err)
	}
}

func TestGenerateRoundSingleElim_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	stageID := 1
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(round_number\), 0\) \+ 1 FROM rounds WHERE stage_id = \$1`).
		WithArgs(stageID).
		WillReturnError(errors.New("db error"))

	err := GenerateRoundSingleElim(db, stageID)
	if err == nil || err.Error() == "" {
		t.Errorf("expected db error, got: %v", err)
	}
}

// --- GenerateRoundDoubleElim ---

func TestGenerateRoundDoubleElim_Success_FirstRound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	stageID := 1

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT COALESCE\(MAX\(round_number\), 0\) \+ 1 FROM rounds WHERE stage_id = \$1 AND bracket = 'W'`).
		WithArgs(stageID).
		WillReturnRows(sqlmock.NewRows([]string{"next_winners_round"}).AddRow(1))

	mock.ExpectQuery(`SELECT COALESCE\(MAX\(round_number\), 0\) \+ 1 FROM rounds WHERE stage_id = \$1 AND bracket = 'L'`).
		WithArgs(stageID).
		WillReturnRows(sqlmock.NewRows([]string{"next_losers_round"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"user_id", "team_id"}).
		AddRow(1, nil).
		AddRow(2, nil)
	mock.ExpectQuery(`SELECT user_id, team_id FROM stage_participants WHERE stage_id=\$1`).
		WithArgs(stageID).
		WillReturnRows(rows)

	mock.ExpectQuery(`INSERT INTO rounds \(stage_id, round_number, bracket\) VALUES \(\$1, \$2, 'W'\) RETURNING round_id`).
		WithArgs(stageID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"round_id"}).AddRow(21))
	mock.ExpectQuery(`INSERT INTO matches \(round_id, scheduled_at\) VALUES \(\$1, NOW\(\)\) RETURNING match_id`).
		WithArgs(21).
		WillReturnRows(sqlmock.NewRows([]string{"match_id"}).AddRow(201))
	mock.ExpectExec(`INSERT INTO match_participants`).WithArgs(
		201, 1, nil, 2, nil,
	).WillReturnResult(sqlmock.NewResult(1, 2))

	mock.ExpectCommit()

	err := GenerateRoundDoubleElim(db, stageID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateRoundDoubleElim_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	stageID := 1
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(round_number\), 0\) \+ 1 FROM rounds WHERE stage_id = \$1 AND bracket = 'W'`).
		WithArgs(stageID).
		WillReturnError(errors.New("db error"))

	err := GenerateRoundDoubleElim(db, stageID)
	if err == nil || err.Error() == "" {
		t.Errorf("expected db error, got: %v", err)
	}
}

// --- GetTopNFromPrevRoundRobin ---

func TestGetTopNFromPrevRoundRobin_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	currentStageID := 2
	prevStageID := 1

	mock.ExpectQuery(`SELECT stage_id FROM competition_stages`).
		WithArgs(currentStageID).
		WillReturnRows(sqlmock.NewRows([]string{"stage_id"}).AddRow(prevStageID))

	mock.ExpectQuery(`SELECT tourney_format_id FROM competition_stages WHERE stage_id = \$1`).
		WithArgs(prevStageID).
		WillReturnRows(sqlmock.NewRows([]string{"tourney_format_id"}).AddRow(3))

	rows := sqlmock.NewRows([]string{"user_id", "team_id"}).
		AddRow(1, nil).
		AddRow(2, nil)
	mock.ExpectQuery(`SELECT user_id, team_id FROM stage_participants WHERE stage_id = \$1`).
		WithArgs(prevStageID).
		WillReturnRows(rows)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM match_participants mp`).
		WithArgs(prevStageID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM match_participants mp`).
		WithArgs(prevStageID, 2).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	top, err := GetTopNFromPrevRoundRobin(db, currentStageID, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(top) != 1 || top[0].UserID == nil || *top[0].UserID != 1 {
		t.Errorf("expected top participant with UserID 1, got %+v", top)
	}
}

func TestGetTopNFromPrevRoundRobin_NoPrevStage(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	currentStageID := 2
	mock.ExpectQuery(`SELECT stage_id FROM competition_stages`).
		WithArgs(currentStageID).
		WillReturnError(sql.ErrNoRows)

	_, err := GetTopNFromPrevRoundRobin(db, currentStageID, 1)
	if err == nil || err.Error() == "" {
		t.Errorf("expected error for no previous stage, got: %v", err)
	}
}

func TestGetTopNFromPrevRoundRobin_NotRoundRobin(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	currentStageID := 2
	prevStageID := 1

	mock.ExpectQuery(`SELECT stage_id FROM competition_stages`).
		WithArgs(currentStageID).
		WillReturnRows(sqlmock.NewRows([]string{"stage_id"}).AddRow(prevStageID))

	mock.ExpectQuery(`SELECT tourney_format_id FROM competition_stages WHERE stage_id = \$1`).
		WithArgs(prevStageID).
		WillReturnRows(sqlmock.NewRows([]string{"tourney_format_id"}).AddRow(2)) // Not round robin

	_, err := GetTopNFromPrevRoundRobin(db, currentStageID, 1)
	if err == nil || err.Error() != "previous stage is not round robin" {
		t.Errorf("expected error for not round robin, got: %v", err)
	}
}

func TestGetTopNFromPrevRoundRobin_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	currentStageID := 2
	prevStageID := 1

	mock.ExpectQuery(`SELECT stage_id FROM competition_stages`).
		WithArgs(currentStageID).
		WillReturnRows(sqlmock.NewRows([]string{"stage_id"}).AddRow(prevStageID))

	mock.ExpectQuery(`SELECT tourney_format_id FROM competition_stages WHERE stage_id = \$1`).
		WithArgs(prevStageID).
		WillReturnRows(sqlmock.NewRows([]string{"tourney_format_id"}).AddRow(3))

	mock.ExpectQuery(`SELECT user_id, team_id FROM stage_participants WHERE stage_id = \$1`).
		WithArgs(prevStageID).
		WillReturnError(errors.New("db error"))

	_, err := GetTopNFromPrevRoundRobin(db, currentStageID, 1)
	if err == nil || err.Error() == "" {
		t.Errorf("expected db error, got: %v", err)
	}
}
