package helpers

import (
	"strings"
)

// NormalizeURL trims whitespace and lowercases a URL
func NormalizeURL(url string) string {
	return strings.TrimSpace(strings.ToLower(url))
}

// IsEmpty checks if a string is empty after trimming spaces
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}
