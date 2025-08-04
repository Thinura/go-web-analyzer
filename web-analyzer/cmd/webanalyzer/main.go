// package main

// import (
// 	"fmt"
// 	"html/template"
// 	"log"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"
// 	"web-analyzer/internal/server"
// )

// func main() {
// 	// Load all templates from the embedded FS
// 	tmpl := template.New("base").Funcs(template.FuncMap{
// 		"upper": strings.ToUpper,
// 		"formatDuration": func(d time.Duration) string {
// 			return fmt.Sprintf("%.2f seconds", d.Seconds())
// 		},
// 	})

// 	entries, err := server.TemplateFS.ReadDir("templates")
// 	if err != nil {
// 		log.Fatalf("failed to read embedded templates: %v", err)
// 	}
// 	for _, e := range entries {
// 		if !e.IsDir() {
// 			name := "templates/" + e.Name()
// 			_, err := tmpl.ParseFS(server.TemplateFS, name)
// 			if err != nil {
// 				log.Fatalf("failed to parse template %s: %v", name, err)
// 			}
// 		}
// 	}

// 	formTmpl := tmpl.Lookup("form.html")
// 	resultTmpl := tmpl.Lookup("result.html")
// 	if formTmpl == nil || resultTmpl == nil {
// 		log.Fatal("form.html or result.html not found in embedded templates")
// 	}

// 	server.SetTemplates(formTmpl, resultTmpl)

// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/", server.ShowForm)
// 	mux.Handle("/api/analyze", server.Chain(
// 		http.HandlerFunc(server.ErrorHandler(server.HandleAnalyzeJSON)),
// 		server.RateLimit,
// 	))
// 	mux.HandleFunc("/result", server.ShowResultPage)

// 	loggedMux := server.LoggingMiddleware(mux)
// 	host := os.Getenv("HOST")
// 	if host == "" {
// 		host = "0.0.0.0"
// 	}
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8080"
// 	}
// 	addr := fmt.Sprintf("%s:%s", host, port)
// 	log.Printf("Server starting on http://%s\n", addr)
// 	if err := http.ListenAndServe(addr, loggedMux); err != nil {
// 		log.Fatal(err)
// 	}
// }

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"web-analyzer/internal/server"
	"web-analyzer/pkg/embed"
)

func main() {
	formTmpl, err := embed.LoadEmbeddedTemplateFile("form.html")
	if err != nil {
		log.Fatalf("Failed to load form.html: %v", err)
	}
	resultTmpl, err := embed.LoadEmbeddedTemplateFile("result.html")
	if err != nil {
		log.Fatalf("Failed to load result.html: %v", err)
	}
	server.SetTemplates(formTmpl, resultTmpl)

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.ShowForm)
	mux.Handle("/api/analyze", server.Chain(
		http.HandlerFunc(server.ErrorHandler(server.HandleAnalyzeJSON)),
		server.RateLimit,
	))
	mux.HandleFunc("/result", server.ShowResultPage)

	loggedMux := server.LoggingMiddleware(mux)
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Server starting on http://%s\n", addr)
	if err := http.ListenAndServe(addr, loggedMux); err != nil {
		log.Fatal(err)
	}
}
