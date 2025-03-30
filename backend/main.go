package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

func enableCORS(next http.Handler) http.Handler {
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

func loginHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "HTTP Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(request.Body).Decode(&creds); err != nil {
		http.Error(writer, "Error reading data", http.StatusBadRequest)
		return
	}

	response := LoginResponse{
		Message: "Login successful",
		Token:   "token-dummy", //Token dummy used as a placeholder
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(response)
}

func connectToDatabase() {
	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		log.Fatal("ERROR: No database connection string")
	}

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("ERROR: Couldn't connect to database:", err)
	}
	defer db.Close()

	var version string
	if err := db.QueryRow("SELECT version()").Scan(&version); err != nil {
		log.Fatal("Failed to execute query:", err)
	}

	log.Printf("PostgreSQL version: %s\n", version)
}

func main() {
	connectToDatabase()

	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)

	handler := enableCORS(mux)

	log.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
