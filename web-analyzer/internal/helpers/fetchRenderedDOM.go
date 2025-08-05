package helpers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var FetchRenderedDOM = func(url string) ([]byte, error) {
	renderServer := os.Getenv("RENDER_SERVER_URL")
	if renderServer == "" {
		renderServer = "http://localhost:3001"
	}

	body := fmt.Sprintf(`{"url": "%s"}`, url)
	req, err := http.NewRequest("POST", renderServer+"/render", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("render server error: %s", b)
	}

	return io.ReadAll(resp.Body)
}
