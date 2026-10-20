package backbone

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"konkour-bench/internal/openrouter"
)

// Mixed routes Claude exclusively to the subscription CLI. There is no paid
// fallback when the CLI is unavailable or does not support a Claude model.
type Mixed struct{ OpenRouter, Claude Client }

func IsMixed(endpoint string) bool  { return strings.HasPrefix(endpoint, "mixed://") }
func ClaudeModel(model string) bool { return strings.HasPrefix(model, "anthropic/claude-") }
func Name(endpoint string) string {
	if IsMixed(endpoint) {
		return "mixed"
	}
	if IsClaude(endpoint) {
		return "claude"
	}
	return "openrouter"
}

// Both underlying endpoints, including the CLI version, are saved in the run.
// Individual request hashes still use the original backend endpoint so their
// completed caches survive switching between single and mixed runs.
func (m *Mixed) Endpoint() string {
	q := url.Values{"openrouter": {m.OpenRouter.Endpoint()}, "claude": {m.Claude.Endpoint()}}
	return "mixed://v1?" + q.Encode()
}
func EndpointFor(endpoint, model string) string {
	if !IsMixed(endpoint) {
		return endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	key := "openrouter"
	if ClaudeModel(model) {
		key = "claude"
	}
	return u.Query().Get(key)
}

func (m *Mixed) Discover(ctx context.Context, selected []string) (openrouter.Catalog, error) {
	groups := map[string][]string{}
	for _, id := range selected {
		key := "openrouter"
		if ClaudeModel(id) {
			key = "claude"
		}
		groups[key] = append(groups[key], id)
	}
	cat := openrouter.Catalog{Models: map[string]openrouter.Model{}, Endpoints: map[string]json.RawMessage{}, FetchedAt: time.Now().UTC()}
	raw := map[string]json.RawMessage{}
	// Authenticate Claude before making any inference or paid-backend request.
	for _, entry := range []struct {
		name   string
		client Client
	}{{"claude", m.Claude}, {"openrouter", m.OpenRouter}} {
		ids := groups[entry.name]
		if len(ids) == 0 {
			continue
		}
		part, err := entry.client.Discover(ctx, ids)
		if err != nil {
			return cat, fmt.Errorf("%s preflight: %w", entry.name, err)
		}
		for _, id := range ids {
			base := strings.TrimSuffix(id, ":exacto")
			if model, ok := part.Models[base]; ok {
				cat.Models[base] = model
			}
			if endpoint, ok := part.Endpoints[base]; ok {
				cat.Endpoints[base] = endpoint
			}
		}
		raw[entry.name] = part.Raw
	}
	var err error
	cat.Raw, err = json.Marshal(raw)
	return cat, err
}
func (m *Mixed) Complete(ctx context.Context, body []byte, save func(openrouter.Progress) error) openrouter.HTTPResult {
	var request struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &request); err != nil || request.Model == "" {
		return openrouter.HTTPResult{Fatal: true, Error: "mixed backend request has no valid model"}
	}
	if ClaudeModel(request.Model) {
		return m.Claude.Complete(ctx, body, save)
	}
	return m.OpenRouter.Complete(ctx, body, save)
}
