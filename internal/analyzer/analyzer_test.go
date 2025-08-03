package analyzer

import (
	"fmt"
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
	html := `<a href="/internal">Internal</a><a href="http://external.com">External</a><a href="::bad">Bad</a><a href="">Empty</a>`
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

func TestToNamedLinks(t *testing.T) {
	links := []string{"http://a.com", "http://a.com", "http://b.com"}
	named := ToNamedLinks(links)

	if len(named) != 2 {
		t.Errorf("Expected 2 unique links, got %d", len(named))
	}

	for _, nl := range named {
		if nl.URL == "http://a.com" && nl.Occurrence != 2 {
			t.Errorf("Expected occurrence 2 for http://a.com, got %d", nl.Occurrence)
		}
		if nl.URL == "http://b.com" && nl.Occurrence != 1 {
			t.Errorf("Expected occurrence 1 for http://b.com, got %d", nl.Occurrence)
		}
	}
}

func TestRelabelDuplicates(t *testing.T) {
	links := []NamedLink{
		{URL: "http://a.com"},
		{URL: "http://a.com"},
		{URL: "http://b.com"},
	}
	labeled := RelabelDuplicates(links)

	expected := map[string]string{
		"http://a.com": "http://a.com (2)",
		"http://b.com": "http://b.com",
	}

	for _, l := range labeled {
		if l.Label != expected[l.URL] {
			t.Errorf("For URL %s, expected label %s, got %s", l.URL, expected[l.URL], l.Label)
		}
	}
}

func TestHTTPError_Error(t *testing.T) {
	err := &HTTPError{StatusCode: 404, Message: "Not Found"}
	if err.Error() != "Not Found" {
		t.Errorf("Expected 'Not Found', got '%s'", err.Error())
	}
}

func TestAnalyzePage_ReadBodyError(t *testing.T) {
	// Simulate a broken response body
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("server does not support hijacking")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatalf("Hijack failed: %v", err)
		}
		conn.Close()
	}))
	defer ts.Close()
	_, err := AnalyzePage(ts.URL)
	if err == nil || !strings.Contains(err.Error(), "failed to fetch") {
		t.Errorf("Expected read body error, got: %v", err)
	}
}

func TestIsLinkAccessible_RequestCreationFails(t *testing.T) {
	var logged string
	logger := func(format string, args ...interface{}) {
		logged = fmt.Sprintf(format, args...)
	}

	// Invalid URL format that causes http.NewRequest to fail
	result := isLinkAccessible("http://[::1]:namedport", 2*time.Second, logger)
	if result != false {
		t.Errorf("Expected false, got %v", result)
	}
	if !strings.Contains(logged, "HEAD request creation failed") {
		t.Errorf("Expected log message for request creation failure, got: %s", logged)
	}
}

func TestAnalyzePage_ConfigLoadFailure(t *testing.T) {
	original := LoadTagConfig
	LoadTagConfig = func(path string) (*TagConfig, error) {
		return nil, fmt.Errorf("simulated config load failure")
	}
	defer func() { LoadTagConfig = original }()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic due to config load failure, but did not panic")
		} else {
			if httpErr, ok := r.(*HTTPError); !ok || !strings.Contains(httpErr.Message, "Failed to load config") {
				t.Errorf("Expected HTTPError with config message, got: %v", r)
			}
		}
	}()

	server := newTestServer("<html><body></body></html>")
	defer server.Close()
	AnalyzePage(server.URL)
}

func TestIsLinkAccessible_RequestFails(t *testing.T) {
	var logged string
	logger := func(format string, args ...interface{}) {
		logged = fmt.Sprintf(format, args...)
	}

	// Use a non-routable IP to simulate a network error
	result := isLinkAccessible("http://10.255.255.1", 1*time.Second, logger)
	if result != false {
		t.Errorf("Expected false, got %v", result)
	}
	if !strings.Contains(logged, "HEAD request failed") {
		t.Errorf("Expected log message for request failure, got: %s", logged)
	}
}
