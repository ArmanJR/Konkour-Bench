package report

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed logos/*.svg
var logoFiles embed.FS

//go:embed logos/LICENSE
var logoLicense string

//go:embed logos.js
var logoScript string

// Bundle the unmodified, pinned artwork as data URLs. Both HTML images and SVG
// chart images use these URLs, including when charts are exported to PNG.
func logoRuntime() (string, error) {
	entries, err := logoFiles.ReadDir("logos")
	if err != nil {
		return "", fmt.Errorf("read embedded logo directory: %w", err)
	}
	logos := make(map[string]string, len(entries))
	for _, entry := range entries {
		body, err := logoFiles.ReadFile("logos/" + entry.Name())
		if err != nil {
			return "", fmt.Errorf("read embedded logo %s: %w", entry.Name(), err)
		}
		logos[strings.TrimSuffix(entry.Name(), ".svg")] = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(body)
	}
	data, err := json.Marshal(logos)
	if err != nil {
		return "", fmt.Errorf("encode embedded logos: %w", err)
	}
	// Keep the upstream license with both standalone HTML and JavaScript exports.
	credit := "// Lobe Icons 1.95.1 — https://github.com/lobehub/lobe-icons"
	for _, line := range strings.Split(strings.TrimSpace(logoLicense), "\n") {
		credit += "\n//"
		if line != "" {
			credit += " " + line
		}
	}
	return credit + "\nconst reportLogos = " + string(data) + ";\n" + logoScript, nil
}
