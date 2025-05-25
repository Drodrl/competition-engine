package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/Drodrl/competition-engine/handlers"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

//go:embed static/*
var staticFiles embed.FS

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

	// Lookup endpoints
	mux.Handle("/api/sports", EnableCORS(handlers.GetSportsHandler(db)))
	mux.Handle("/api/structure-types", EnableCORS(handlers.GetStructureTypesHandler(db)))
	mux.Handle("/api/tourney-formats", EnableCORS(handlers.GetTournamentFormatsHandler(db)))

	// Static files (Angular app)
	mux.Handle("/", fileServer)

	log.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", EnableCORS(mux)))
}
