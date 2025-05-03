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
