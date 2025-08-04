package embed

import (
	"embed"
	"fmt"
	"html/template"
	"strings"
	"time"
)

//go:embed templates/*.html
var TemplateFS embed.FS

//go:embed config/*.json
var ConfigFS embed.FS

// LoadEmbeddedTemplateFile parses a specific template file from embedded FS with default FuncMap
func LoadEmbeddedTemplateFile(filename string) (*template.Template, error) {
	tmpl := template.New(filename).Funcs(template.FuncMap{
		"upper": strings.ToUpper,
		"formatDuration": func(d time.Duration) string {
			return fmt.Sprintf("%.2f seconds", d.Seconds())
		},
	})
	return tmpl.ParseFS(TemplateFS, "templates/"+filename)
}

// LoadEmbeddedConfigFile reads a JSON config file from embedded FS
func LoadEmbeddedConfigFile(filename string) ([]byte, error) {
	data, err := ConfigFS.ReadFile("config/" + filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}
	return data, nil
}
