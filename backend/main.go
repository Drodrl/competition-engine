<<<<<<< Updated upstream
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	//"github.com/Drodrl/competition-engine/handlers"
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

	//handlers.SetDB(db)

	staticContent, err := fs.Sub(staticFiles, "static/browser")
	if err != nil {
		log.Fatalf("failed to create sub FS: %v", err)
	}

	fileServer := http.FileServer(http.FS(staticContent))

	mux := http.NewServeMux()
	mux.Handle("/login", EnableCORS(NewLoginHandler(db)))

	//mux.Handle("/api/competitions", EnableCORS(handlers.CreateFullCompetitionHandler(db)))
	//mux.Handle("/api/competitions/draft", EnableCORS(http.HandlerFunc(handlers.CreateDraftCompetition)))
	//mux.Handle("/api/competitions/organizer/", EnableCORS(http.HandlerFunc(handlers.GetCompetitionsByOrganizer)))
	//mux.Handle("/api/competitions/", EnableCORS(handlers.CompetitionByIDHandler()))

	mux.Handle("/", fileServer)

	//mux.Handle("/api/sports", EnableCORS(handlers.GetSportsHandler(db)))
	//mux.Handle("/api/structure-types", EnableCORS(handlers.GetStructureTypesHandler(db)))
	//mux.Handle("/api/activity-types", EnableCORS(handlers.GetActivityTypesHandler(db)))
	//mux.Handle("/api/tourney-formats", EnableCORS(handlers.GetTournamentFormatsHandler(db)))

	log.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", EnableCORS(mux)))
}
=======
package main

import (
	"log"
	"net/http"
)

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

	mux := http.NewServeMux()
	mux.Handle("/login", EnableCORS(NewLoginHandler(db)))

	log.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", EnableCORS(mux)))
}
>>>>>>> Stashed changes
