package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "¡Hello world! This app is deployed with Render and CI/CD.")
}

func main() {
	http.HandleFunc("/", handler)

	port := "8080"
	log.Printf("Server listening on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}