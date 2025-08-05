package analyzer

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

func RenderDOMViaPuppeteer(url string) ([]byte, error) {
	renderURL := os.Getenv("RENDER_SERVER_URL")
	if renderURL == "" {
		renderURL = "http://localhost:3001" // fallback
	}

	payload := fmt.Sprintf(`{"url":"%s"}`, url)
	req, err := http.NewRequest("POST", renderURL+"/render", bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call render server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("render server returned %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
