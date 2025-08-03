package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
	"web-analyzer/internal/server"
)

// func main() {
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/", server.ShowForm)

// 	mux.Handle("/api/analyze", server.Chain(
// 		http.HandlerFunc(server.ErrorHandler(server.HandleAnalyzeJSON)),
// 		server.RateLimit,
// 	))
// 	mux.HandleFunc("/result", server.ShowResultPage)

// 	// Wrap with middleware
// 	loggedMux := server.LoggingMiddleware(mux)

// 	log.Println("Server starting on http://localhost:8080")
// 	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
// 		log.Fatal(err)
// 	}
// }

func main() {
	// Parse templates at startup
	formTmpl, err := template.ParseFiles("internal/server/templates/form.html")
	if err != nil {
		log.Fatalf("Failed to load form.html: %v", err)
	}
	resultTmpl, err := template.New("result.html").
		Funcs(template.FuncMap{
			"upper": strings.ToUpper,
			"formatDuration": func(d time.Duration) string {
				return fmt.Sprintf("%.2f seconds", d.Seconds())
			},
		}).
		ParseFiles("internal/server/templates/result.html")
	if err != nil {
		log.Fatalf("Failed to load result.html: %v", err)
	}

	// Assign to handlers
	server.SetTemplates(formTmpl, resultTmpl)

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.ShowForm)
	mux.Handle("/api/analyze", server.Chain(
		http.HandlerFunc(server.ErrorHandler(server.HandleAnalyzeJSON)),
		server.RateLimit,
	))
	mux.HandleFunc("/result", server.ShowResultPage)

	loggedMux := server.LoggingMiddleware(mux)
	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatal(err)
	}
}
