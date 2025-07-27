package analyzer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer spins up an HTTP test server that returns the provided HTML.
func newTestServer(html string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
}

// assertEqual is a generic assertion helper for comparing two values.
func assertEqual(t *testing.T, name string, got, want interface{}) {
	t.Helper()
	if got != want {
		t.Errorf("Mismatch on %s: expected %v, got %v", name, want, got)
	}
}

// Predefined reusable HTML snippets for tests.
const basicLoginFormHTML = `
	<!DOCTYPE html>
	<html>
	<head><title>Test Login</title></head>
	<body>
		<h1>Login</h1>
		<form>
			<input type="text" name="username" />
			<input type="password" name="password" />
		</form>
	</body>
	</html>
`

const basicNonLoginHTML = `
	<!DOCTYPE html>
	<html>
	<head><title>Simple Page</title></head>
	<body>
		<h1>Hello</h1>
		<p>This page has no login form.</p>
	</body>
	</html>
`

const basicTestHTML = `
		<!DOCTYPE html>
		<html>
		<head><title>Test Page</title></head>
		<body>
			<h1>Main Heading</h1>
			<h2>Sub Heading</h2>
			<a href="/internal">Internal</a>
			<a href="https://external.com">External</a>
			<form><input type="password" /></form>
		</body>
		</html>
	`

func assertNamedLinksCount(t *testing.T, name string, got []NamedLink, want int) {
	t.Helper()
	if len(got) != want {
		t.Errorf("Expected %s count %d, got %d", name, want, len(got))
	}
}
