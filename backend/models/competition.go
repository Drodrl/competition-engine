package models

import "time"

type Competition struct {
	CompetitionId   int     `json:"competition_id"`
	CompetitionName string  `json:"competition_name"`
	SportID         int     `json:"sport_id"`
	SportName       string  `json:"sport_name"`
	StartDate       *string `json:"start_date"`
	EndDate         *string `json:"end_date"`
	DateCreated     *string `json:"date_created"`
	DateUpdated     *string `json:"date_updated"`
	OrganizerID     int     `json:"organizer_id"`
	Status          int     `json:"status"`
	MaxParticipants *int    `json:"max_participants"`
	FlagTeams       bool    `json:"flag_teams"`
}

type StageDTO struct {
	StageID             int    `json:"stage_id"`
	StageName           string `json:"stage_name"`
	StageOrder          int    `json:"stage_order"`
	TourneyFormatID     int    `json:"tourney_format_id"`
	ParticipantsAtStart int    `json:"participants_at_start"`
	ParticipantsAtEnd   int    `json:"participants_at_end"`
}

type StageRound struct {
	RoundID     int `json:"round_id"`
	StageID     int `json:"stage_id"`
	RoundNumber int `json:"round_number"`
}

type Match struct {
	MatchID     int        `json:"match_id"`
	RoundID     int        `json:"round_id"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type MatchParticipant struct {
	MatchID  int  `json:"match_id"`
	UserID   *int `json:"user_id"`
	TeamID   *int `json:"team_id"`
	IsWinner bool `json:"is_winner"`
	Score    *int `json:"score"`
}
