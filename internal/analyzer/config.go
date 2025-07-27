// config.go
package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
)

type TagConfig struct {
	Headings []string `json:"headings"`
}

// Declare LoadTagConfigFunc as a variable that can be overridden in tests
var LoadTagConfig = func(path string) (*TagConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg TagConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config JSON: %w", err)
	}
	return &cfg, nil
}
