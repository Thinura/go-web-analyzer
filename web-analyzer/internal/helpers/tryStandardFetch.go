package helpers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"web-analyzer/pkg/errors"
)

func TryStandardFetch(url string) ([]byte, bool, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, false, &errors.HTTPError{StatusCode: http.StatusInternalServerError, Message: fmt.Sprintf("failed to fetch: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		// Likely a redirect to bot-check or login
		return nil, true, nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, &errors.HTTPError{StatusCode: http.StatusInternalServerError, Message: fmt.Sprintf("failed to read body: %v", err)}
	}

	// Check for common bot-block HTML signs
	if strings.Contains(strings.ToLower(string(data)), "captcha") || strings.Contains(string(data), "window._cf_chl_opt") {
		return data, true, nil
	}

	return data, false, nil
}
