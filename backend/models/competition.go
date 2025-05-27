package models

type Competition struct {
	CompetitionId   int     `json:"competition_id"`
	CompetitionName string  `json:"competition_name"`
	SportID         int     `json:"sport_id"`
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
