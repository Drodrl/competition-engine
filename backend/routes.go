package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/Drodrl/competition-engine/handlers"
	"github.com/gorilla/mux"
)

//go:embed static/*
var staticFiles embed.FS

func NewRouter(db *sql.DB) http.Handler {
	handlers.SetDB(db)
	router := mux.NewRouter()

	// --- Auth ---
	router.Handle("/login", EnableCORS(NewLoginHandler(db))).Methods("POST")

	// --- Competition Management ---
	router.Handle("/api/competitions", EnableCORS(http.HandlerFunc(handlers.GetAllCompetitions))).Methods("GET")
	router.Handle("/api/competitions/draft", EnableCORS(http.HandlerFunc(handlers.CreateDraftCompetition))).Methods("POST")
	router.Handle("/api/competitions/organizer/{organizerId}", EnableCORS(http.HandlerFunc(handlers.GetCompetitionsByOrganizer))).Methods("GET")
	router.Handle("/api/competitions/{competitionId}", EnableCORS(handlers.CompetitionByIDHandler())).Methods("GET", "DELETE", "PUT")
	router.Handle("/api/competitions/{competitionId}/status", EnableCORS(handlers.CompetitionByIDHandler())).Methods("PATCH")
	router.Handle("/api/competitions/{competitionId}/participants", EnableCORS(http.HandlerFunc(handlers.GetParticipantsByCompetitionID))).Methods("GET")
	router.Handle("/api/competitions/{competitionId}/finish", EnableCORS(http.HandlerFunc(handlers.FinishCompetition))).Methods("POST")
	router.Handle("/api/competitions/flag_teams/{flagTeams}", EnableCORS(http.HandlerFunc(handlers.GetCompetitionsByFlagTeams))).Methods("GET")

	// --- Competition Stages ---
	router.Handle("/api/competitions/{competitionId}/stages", EnableCORS(http.HandlerFunc(handlers.GetStagesByCompetitionID))).Methods("GET")
	router.Handle("/api/competitions/{competitionId}/stages", EnableCORS(http.HandlerFunc(handlers.AddStageToCompetition))).Methods("POST")
	router.Handle("/api/competitions/{competitionId}/stages/{stageId}", EnableCORS(http.HandlerFunc(handlers.UpdateStage))).Methods("PUT")
	router.Handle("/api/competitions/{competitionId}/stages/{stageId}", EnableCORS(http.HandlerFunc(handlers.DeleteStage))).Methods("DELETE")

	// --- Rounds ---
	router.Handle("/api/stages/{stageId}/rounds", EnableCORS(http.HandlerFunc(handlers.GetRoundsByStageID))).Methods("GET")
	router.Handle("/api/stages/{stageId}/generate-next-round", EnableCORS(http.HandlerFunc(handlers.GenerateNextRound))).Methods("POST")
	router.Handle("/api/stages/{stageId}/can-generate-next-round", EnableCORS(http.HandlerFunc(handlers.CanGenerateNextRound))).Methods("GET")
	router.Handle("/api/stages/{stageId}/advance", EnableCORS(http.HandlerFunc(handlers.AdvanceAfterRoundRobin))).Methods("POST")

	// --- Matches ---
	router.Handle("/api/rounds/{roundId}/matches", EnableCORS(http.HandlerFunc(handlers.GetMatchesByRoundID))).Methods("GET")
	router.Handle("/api/matches/{matchId}/participants", EnableCORS(http.HandlerFunc(handlers.GetMatchParticipants))).Methods("GET")
	router.Handle("/api/matches/{matchId}/participants", EnableCORS(http.HandlerFunc(handlers.UpdateMatchResult))).Methods("PUT")
	router.Handle("/api/matches/{matchId}/results", EnableCORS(http.HandlerFunc(handlers.SaveMatchResults))).Methods("PUT")

	// --- Lookup/Reference Data ---
	router.Handle("/api/sports", EnableCORS(handlers.GetSportsHandler(db))).Methods("GET")
	router.Handle("/api/structure-types", EnableCORS(handlers.GetStructureTypesHandler(db))).Methods("GET")
	router.Handle("/api/tourney-formats", EnableCORS(handlers.GetTournamentFormatsHandler(db))).Methods("GET")

	// --- Team & User Management ---
	router.Handle("/api/user-teams", EnableCORS(handlers.GetUserTeamsHandler(db)))
	router.Handle("/api/team-participants", EnableCORS(handlers.GetTeamParticipantsHandler(db)))
	router.Handle("/api/remove-participants", EnableCORS(handlers.RemoveParticipantsHandler(db)))
	router.Handle("/api/add-participants", EnableCORS(handlers.AddParticipantsHandler(db)))

	// --- Other Handlers ---
	router.Handle("/api/handlers/athletes", EnableCORS(handlers.NewAthletesHandler(db))).Methods("GET", "POST")
	router.Handle("/api/handlers/teams", EnableCORS(handlers.NewTeamsHandler(db))).Methods("GET", "POST")
	router.Handle("/handlers/user_signup", EnableCORS(handlers.NewUserSignupHandler(db))).Methods("POST")
	router.Handle("/handlers/team_signup", EnableCORS(handlers.NewTeamSignupHandler(db))).Methods("POST")
	router.Handle("/handlers/team_create", EnableCORS(handlers.NewTeamCreateHandler(db))).Methods("POST")
	router.Handle("/api/handlers/competitions", EnableCORS(handlers.NewUserSignupHandler(db))) // If needed

	// --- Static Files ---
	staticContent, err := fs.Sub(staticFiles, "static/browser")
	if err != nil {
		log.Fatalf("failed to create sub FS: %v", err)
	}
	staticFS := http.FileServer(http.FS(staticContent))
	router.PathPrefix("/").Handler(staticFS)

	return router
}
