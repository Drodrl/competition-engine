package models

type Competition struct {
	CompetitionName string     `json:"competition_name"`
	SportID         int        `json:"sport_id"`
	StartDate       string     `json:"start_date"`
	EndDate         string     `json:"end_date"`
	OrganizerID     int        `json:"organizer_id"`
	StructureTypeID int        `json:"structure_type_id"`
	Stages          []StageDTO `json:"stages"`
}

type StageDTO struct {
	StageName       string `json:"stage_name"`
	StageOrder      int    `json:"stage_order"`
	TourneyFormatID int    `json:"tourney_format_id"`
	ActivityTypeID  int    `json:"activity_type_id"`
	Participants    int    `json:"participants"`
}
