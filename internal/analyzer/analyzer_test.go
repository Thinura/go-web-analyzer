// analyzer_test.go contains tests for AnalyzePage() and full-page analysis logic.
// Helpers are defined in helper_test.go for reuse across this package.
package analyzer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func init() {
	LoadTagConfig = func(path string) (*TagConfig, error) {
		return &TagConfig{Headings: []string{"h1", "h2", "h3"}}, nil
	}
}

func TestAnalyzePage_BasicPage(t *testing.T) {
	server := newTestServer(basicTestHTML)

	defer server.Close()

	result, err := AnalyzePage(server.URL)
	if err != nil {
		t.Fatalf("AnalyzePage failed: %v", err)
	}

	assertEqual(t, "Title", result.Title, "Test Page")
	assertNamedLinksCount(t, "InternalLinks", result.InternalLinks, 1)
	assertNamedLinksCount(t, "ExternalLinks", result.ExternalLinks, 1)
	assertEqual(t, "HeadingsCount[h1]", result.HeadingsCount["h1"], 1)
	assertEqual(t, "HeadingsCount[h2]", result.HeadingsCount["h2"], 1)
	assertEqual(t, "HasLoginForm", result.HasLoginForm, true)
}

func TestAnalyzePage_InvalidURL(t *testing.T) {
	_, err := AnalyzePage("://bad-url")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestAnalyzePage_Non200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	_, err := AnalyzePage(server.URL)
	if err == nil || !strings.Contains(err.Error(), "HTTP error") {
		t.Errorf("Expected HTTP error, got: %v", err)
	}
}

func TestAnalyzePage_BadBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, _ := w.(http.Hijacker)
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer server.Close()

	_, err := AnalyzePage(server.URL)
	if err == nil || !strings.Contains(err.Error(), "failed to fetch") {
		t.Errorf("Expected fetch failure, got: %v", err)
	}
}

func TestAnalyzePage_BadHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><title>Broken"))
	}))
	defer server.Close()

	_, err := AnalyzePage(server.URL)
	if err != nil {
		t.Errorf("Expected recoverable HTML parse, got: %v", err)
	}
}

func TestAnalyzePage_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(""))
	}))
	defer server.Close()

	result, err := AnalyzePage(server.URL)
	if err != nil {
		t.Fatalf("AnalyzePage failed: %v", err)
	}

	if result.HTMLVersion != "Unknown or Custom DOCTYPE" {
		t.Errorf("Expected 'Unknown or Custom DOCTYPE', got '%s'", result.HTMLVersion)
	}
}

func TestClassifyLinksConcurrently(t *testing.T) {
	links := []NamedLink{
		{URL: "https://example.com", Label: "Example"},
		{URL: "https://nonexistent.abcxyz", Label: "Fake"},
	}
	config := LinkCheckerConfig{
		MaxConcurrency: 2,
		Timeout:        2 * time.Second,
	}

	accessible, inaccessible := ClassifyLinksConcurrently(links, config)

	if len(accessible)+len(inaccessible) != 2 {
		t.Errorf("Expected 2 total results, got %d", len(accessible)+len(inaccessible))
	}
}

func TestAnalyzePage_CustomHeadingTags(t *testing.T) {
	html := `<html><body><custom-heading>Custom Title</custom-heading></body></html>`
	ts := newTestServer(html)
	defer ts.Close()

	// Temporarily override LoadTagConfig
	orig := LoadTagConfig
	LoadTagConfig = func(path string) (*TagConfig, error) {
		return &TagConfig{Headings: []string{"custom-heading"}}, nil
	}
	defer func() { LoadTagConfig = orig }()

	result, err := AnalyzePage(ts.URL)
	if err != nil {
		t.Fatalf("AnalyzePage failed: %v", err)
	}

	if len(result.Headings) != 1 || result.Headings[0].Tag != "custom-heading" {
		t.Errorf("Expected one custom-heading, got: %+v", result.Headings)
	}
}

func TestAnalyzePage_LinkClassification(t *testing.T) {
	html := `<a href="/internal">Internal</a><a href="http://external.com">External</a>`
	ts := newTestServer(html)
	defer ts.Close()

	result, err := AnalyzePage(ts.URL)
	if err != nil {
		t.Fatalf("AnalyzePage failed: %v", err)
	}

	if len(result.InternalLinks) != 1 || len(result.ExternalLinks) != 1 {
		t.Errorf("Expected 1 internal and 1 external link, got %d internal, %d external", len(result.InternalLinks), len(result.ExternalLinks))
	}
}

func TestAnalyzePage_LoginFormDetection(t *testing.T) {
	html := `<form><input type="password" /></form>`
	ts := newTestServer(html)
	defer ts.Close()

	result, err := AnalyzePage(ts.URL)
	if err != nil {
		t.Fatalf("AnalyzePage failed: %v", err)
	}

	if !result.HasLoginForm {
		t.Errorf("Expected login form detection")
	}
}

func TestDeduplicateLinks(t *testing.T) {
	links := []string{"http://a.com", "http://a.com", "http://b.com"}
	deduped := deduplicateLinks(links)

	if len(deduped) != 3 {
		t.Errorf("Expected 3 labeled links, got %d", len(deduped))
	}

	if deduped[1].Label != "http://a.com (duplicate 2)" {
		t.Errorf("Expected duplicate label, got %s", deduped[1].Label)
	}
}

func TestDetectHTMLVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<!DOCTYPE html>", "HTML5"},
		{"<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\">", "HTML 4.01 Transitional"},
		{"<!DOCTYPE html><html>", "HTML5"},
		{"<html>", "Unknown or Custom DOCTYPE"},
	}
	for _, tt := range tests {
		version := detectHTMLVersion([]byte(tt.input))
		if version != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, version)
		}
	}
}
