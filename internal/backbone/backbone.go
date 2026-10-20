// Package backbone defines the inference boundary. Backends normalize results
// into the existing generation/usage envelope used by storage and scoring.
package backbone

import (
	"context"
	"fmt"
	"strings"

	"konkour-bench/internal/claude"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
)

type Client interface {
	Endpoint() string
	Discover(context.Context, []string) (openrouter.Catalog, error)
	Complete(context.Context, []byte, func(openrouter.Progress) error) openrouter.HTTPResult
}

func IsClaude(endpoint string) bool { return strings.HasPrefix(endpoint, "claude-cli://") }

func Build(endpoint string, q dataset.Question, model, effort string, stream bool) ([]byte, error) {
	endpoint = EndpointFor(endpoint, model)
	if endpoint == "" {
		return nil, fmt.Errorf("missing backend endpoint for %s", model)
	}
	if IsClaude(endpoint) {
		return claude.Build(q, model, effort)
	}
	if stream {
		return openrouter.BuildStream(q, model, effort)
	}
	return openrouter.Build(q, model, effort)
}
