package analyzer

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
