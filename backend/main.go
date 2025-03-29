package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Credentials struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Token string `json:"token,omitempty"` 
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)

	handler := enableCORS(mux)

	log.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}