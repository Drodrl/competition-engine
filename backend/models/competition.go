package models

type Competition struct {
	ID        int    `json:"competition_id"`
	Name      string `json:"competition_name"`
	Sport     string `json:"sport_id"`
	StartDate string `json:"start_date"`
}
