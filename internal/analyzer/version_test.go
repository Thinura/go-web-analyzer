package analyzer

import (
	"testing"
)

func TestDetectHTMLVersion_KnownDoctypes(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{"HTML5", "<!DOCTYPE html><html><head></head><body></body></html>", "HTML5"},
		{"HTML 4.01 Transitional", `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">`, "HTML 4.01 Transitional"},
		{"HTML 4.01 Strict", `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN">`, "HTML 4.01 Strict"},
		{"XHTML Strict", `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN">`, "XHTML 1.0 Strict"},
		{"XHTML Transitional", `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN">`, "XHTML 1.0 Transitional"},
		{"Unknown", "<html><body>No doctype</body></html>", "Unknown or Custom DOCTYPE"},
		{"Unknown (regex fallback)", "<!DOCTYPE WeirdHTML SYSTEM 'x.dtd'>", "Unknown DOCTYPE: weirdhtml system 'x.dtd'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := detectHTMLVersion([]byte(tt.html))
			if actual != tt.expected {
				t.Errorf("Expected %s, got %s\nHTML: %s", tt.expected, actual, tt.html)
			}
		})
	}
}
