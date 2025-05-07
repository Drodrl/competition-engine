package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Drodrl/competition-engine/models"
)

func CreateFullCompetitionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var competition models.Competition

		if err := json.NewDecoder(r.Body).Decode(&competition); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if competition.StructureTypeID == models.SingleInstance && len(competition.Stages) != 1 {
			http.Error(w, "Single instance competitions must have exactly one stage", http.StatusBadRequest)
			return
		}

		if len(competition.Stages) > 3 {
			http.Error(w, "Competitions can have a maximum of 3 stages", http.StatusBadRequest)
			return
		}

		lastFormat := competition.Stages[len(competition.Stages)-1].TourneyFormatID
		if lastFormat == models.Groups {
			http.Error(w, "Groups cannot be the last stage", http.StatusBadRequest)
			return
		}

		var competitionID int
		err := db.QueryRow(`
		INSERT INTO competitions (competition_name, sport_id, start_date, end_date, organizer_id, structure_type_id)
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING competition_id
		`, competition.CompetitionName, competition.SportID, competition.StartDate, competition.EndDate, competition.OrganizerID, competition.StructureTypeID).Scan(&competitionID)

		if err != nil {
			http.Error(w, "Failed to create competition", http.StatusInternalServerError)
			return
		}

		for _, stage := range competition.Stages {
			_, err := db.Exec(`
    			INSERT INTO competition_stages 
    			(competition_id, stage_order, stage_name, tourney_format_id, activity_type_id, participants)
    			VALUES ($1, $2, $3, $4, $5, $6)
			`, competitionID, stage.StageOrder, stage.StageName, stage.TourneyFormatID, stage.ActivityTypeID, stage.Participants)
			if err != nil {
				http.Error(w, "Error inserting stages", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"competition_id": competitionID})

	}
}
