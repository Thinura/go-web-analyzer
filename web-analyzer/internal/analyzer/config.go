package analyzer

import (
	"encoding/json"
	"fmt"
	"web-analyzer/pkg/embed"
)

type TagConfig struct {
	Headings    []string `json:"headings"`
	AllowedTags []string `json:"allowedTags"`
}

var LoadTagConfig = func() (*TagConfig, error) {
	data, err := embed.LoadEmbeddedConfigFile("config.json")
	if err != nil {
		return nil, err
	}
	var cfg TagConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}
