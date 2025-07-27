package main

import (
	"log"
	"net/http"
	"web-analyzer/internal/server"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.ShowForm)
	mux.HandleFunc("/analyze", server.HandleAnalyze)
	mux.Handle("/api/analyze", server.Chain(
		http.HandlerFunc(server.HandleAnalyzeJSON),
		server.RateLimit,
	))
	mux.HandleFunc("/result", server.ShowResultPage)

	// Wrap with middleware
	loggedMux := server.LoggingMiddleware(mux)

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatal(err)
	}
}
