package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/Drodrl/competition-engine/handlers"
)

//go:embed static/*
var staticFiles embed.FS

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	db, err := OpenDatabase()
	if err != nil {
		log.Fatal("ERROR: Couldn't open database:", err)
	}
	defer db.Close()

	handlers.SetDB(db)

	staticContent, err := fs.Sub(staticFiles, "static/browser")
	if err != nil {
		log.Fatalf("failed to create sub FS: %v", err)
	}

	fileServer := http.FileServer(http.FS(staticContent))

	mux := http.NewServeMux()
	mux.Handle("/login", EnableCORS(NewLoginHandler(db)))
	// Competition endpoints
	mux.Handle("/api/competitions/draft", EnableCORS(http.HandlerFunc(handlers.CreateDraftCompetition)))
	mux.Handle("/api/competitions/organizer/", EnableCORS(http.HandlerFunc(handlers.GetCompetitionsByOrganizer)))
	mux.Handle("/api/competitions/", EnableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/stages") && r.Method == http.MethodGet:
			handlers.GetStagesByCompetitionID(w, r)
		case strings.HasSuffix(path, "/stages") && r.Method == http.MethodPost:
			handlers.AddStageToCompetition(w, r)
		case strings.Contains(path, "/stages/") && r.Method == http.MethodPut:
			handlers.UpdateStage(w, r)
		case strings.Contains(path, "/stages/") && r.Method == http.MethodDelete:
			handlers.DeleteStage(w, r)
		case r.Method == http.MethodDelete:
			handlers.DeleteCompetition(w, r)
		default:
			handlers.CompetitionByIDHandler().ServeHTTP(w, r)
		}
	})))
	mux.Handle("/api/competitions", EnableCORS(http.HandlerFunc(handlers.GetAllCompetitions)))
	// mux.Handle("/api/handlers/competitions", EnableCORS(handlers.NewCompetitionListHandler(db)))
	mux.Handle("/api/handlers/competitions", EnableCORS(handlers.NewUserSignupHandler(db)))
	mux.Handle("/api/competitions/flag_teams/", EnableCORS(http.HandlerFunc(handlers.GetCompetitionsByFlagTeams)))
	mux.Handle("/api/handlers/athletes", EnableCORS(handlers.NewAthletesHandler(db)))
	mux.Handle("/api/handlers/teams", EnableCORS(handlers.NewTeamsHandler(db)))
	mux.Handle("/handlers/user_signup", EnableCORS(handlers.NewUserSignupHandler(db)))
	mux.Handle("/handlers/team_signup", EnableCORS(handlers.NewTeamSignupHandler(db)))
	mux.Handle("/handlers/team_create", EnableCORS(handlers.NewTeamCreateHandler(db)))

	// Lookup endpoints
	mux.Handle("/api/sports", EnableCORS(handlers.GetSportsHandler(db)))
	mux.Handle("/api/structure-types", EnableCORS(handlers.GetStructureTypesHandler(db)))
	mux.Handle("/api/tourney-formats", EnableCORS(handlers.GetTournamentFormatsHandler(db)))
	mux.Handle("/api/user-teams", EnableCORS(handlers.GetUserTeamsHandler(db)))
	mux.Handle("/api/team-participants", EnableCORS(handlers.GetTeamParticipantsHandler(db)))

	// Teams endpoints
	mux.Handle("/api/remove-participants", EnableCORS(handlers.RemoveParticipantsHandler(db)))
	mux.Handle("/api/add-participants", EnableCORS(handlers.AddParticipantsHandler(db)))

	// Static files (Angular app)
	mux.Handle("/", fileServer)

	log.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", EnableCORS(mux)))
}
