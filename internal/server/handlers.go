package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
	"web-analyzer/internal/analyzer"
)

var (
	formTmpl   = template.Must(template.ParseFiles("internal/server/templates/form.html"))
	resultTmpl = template.Must(template.New("result.html").
			Funcs(template.FuncMap{
			"upper": strings.ToUpper,
			"formatDuration": func(d time.Duration) string {
				return fmt.Sprintf("%.2f seconds", d.Seconds())
			},
		}).
		ParseFiles("internal/server/templates/result.html"))
)

// ShowForm serves the input form.
func ShowForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := formTmpl.Execute(w, nil); err != nil {
		http.Error(w, "Failed to render form: "+err.Error(), http.StatusInternalServerError)
	}
}

// HandleAnalyze processes the form and shows results.
func HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pageURL := r.FormValue("url")
	if pageURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	result, err := analyzer.AnalyzePage(pageURL)
	if err != nil {
		http.Error(w, "Failed to analyze: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Configure concurrent link checking
	config := analyzer.LinkCheckerConfig{
		MaxConcurrency: 10,
		Timeout:        5 * time.Second,
		Logger:         func(format string, args ...interface{}) {},
	}

	// Merge internal and external links into one slice of NamedLink
	allLinks := make([]analyzer.NamedLink, 0, len(result.InternalLinks)+len(result.ExternalLinks))
	allLinks = append(allLinks, result.InternalLinks...)
	allLinks = append(allLinks, result.ExternalLinks...)

	result.AccessibleLinks, result.InaccessibleLinks = analyzer.ClassifyLinksConcurrently(allLinks, config)

	if err := resultTmpl.Execute(w, result); err != nil {
		http.Error(w, "Failed to render result: "+err.Error(), http.StatusInternalServerError)
	}
}

// HandleAnalyzeJSON processes the URL and returns JSON
func HandleAnalyzeJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	pageURL := r.FormValue("url")
	if pageURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if cached, ok := analyzer.GetFromCache(pageURL); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}

	// Start analysis timer
	start := time.Now()

	// Run analysis
	result, err := analyzer.AnalyzePage(pageURL)
	if err != nil {
		http.Error(w, "Failed to analyze: "+err.Error(), http.StatusBadRequest)
		return
	}
	result.AnalysisDuration = time.Since(start)

	// Link classification
	config := analyzer.LinkCheckerConfig{
		MaxConcurrency: 10,
		Timeout:        5 * time.Second,
	}
	result.AccessibleLinks, result.InaccessibleLinks = analyzer.ClassifyLinksConcurrently(
		append(result.InternalLinks, result.ExternalLinks...), config,
	)

	// Store in cache
	analyzer.StoreInCache(pageURL, result)

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode JSON: "+err.Error(), http.StatusInternalServerError)
	}
}

func ShowResultPage(w http.ResponseWriter, r *http.Request) {
	// Serve the result.html template shell
	tmpl := template.Must(template.ParseFiles("internal/server/templates/result.html"))
	tmpl.Execute(w, nil)
}
