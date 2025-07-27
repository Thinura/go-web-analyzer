package analyzer

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckLinksConcurrently_Mixed(t *testing.T) {
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer okServer.Close()

	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone", http.StatusGone)
	}))
	defer failServer.Close()

	links := []string{
		okServer.URL,
		failServer.URL,
		"http://127.0.0.1:1", // definitely unreachable
	}

	config := LinkCheckerConfig{
		MaxConcurrency: 3,
		Timeout:        2 * time.Second,
	}

	count := checkLinksConcurrently(links, config)
	if count != 2 {
		t.Errorf("Expected 2 inaccessible links, got %d", count)
	}
}

func TestIsLinkAccessible_ValidAndInvalid(t *testing.T) {
	// ✅ Working server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ok := isLinkAccessible(server.URL, 2*time.Second, nil)
	if !ok {
		t.Errorf("Expected accessible link to return true")
	}

	// ❌ Broken server
	badURL := "http://127.0.0.1:1"
	ok = isLinkAccessible(badURL, 2*time.Second, nil)
	if ok {
		t.Errorf("Expected inaccessible link to return false")
	}
}
