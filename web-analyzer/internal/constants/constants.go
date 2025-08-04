package constants

import "time"

// Timeouts used across the application
const (
	// RequestTimeout defines the timeout for HTTP requests to fetch pages.
	RequestTimeout = 10 * time.Second

	// LinkCheckTimeout defines the timeout for checking if a link is accessible.
	LinkCheckTimeout = 5 * time.Second
)

// DefaultHTMLVersion is used when no DOCTYPE is explicitly detected.
const DefaultHTMLVersion = "HTML5 (assumed)"
