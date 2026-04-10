package oauth

import (
	"embed"
	"fmt"
)

//go:embed templates/*.html
var templateFS embed.FS

func loadTemplate(name string) string {
	raw, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		return fmt.Sprintf("<html><body><h1>%s</h1></body></html>", name)
	}
	return string(raw)
}
