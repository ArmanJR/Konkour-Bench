package backbone

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
)

type routeFixture struct {
	endpoint string
	selected []string
	calls    int
	fail     bool
}

func (c *routeFixture) Endpoint() string { return c.endpoint }
func (c *routeFixture) Discover(_ context.Context, ids []string) (openrouter.Catalog, error) {
	c.selected = ids
	if c.fail {
		return openrouter.Catalog{}, errors.New("subscription unavailable")
	}
	models := map[string]openrouter.Model{}
	for _, id := range ids {
		models[id] = openrouter.Model{ID: id}
	}
	return openrouter.Catalog{Models: models, Raw: json.RawMessage(`{}`)}, nil
}
func (c *routeFixture) Complete(context.Context, []byte, func(openrouter.Progress) error) openrouter.HTTPResult {
	c.calls++
	return openrouter.HTTPResult{Status: 200}
}

func TestMixedRoutesAndPreservesRequestIdentity(t *testing.T) {
	api := &routeFixture{endpoint: "https://gateway.test/api/v1"}
	cli := &routeFixture{endpoint: "claude-cli://v1/2.1.296"}
	m := Mixed{OpenRouter: api, Claude: cli}
	ids := []string{"anthropic/claude-sonnet-5.5", "openai/gpt-6-luna", "anthropic/claude-opus-5.5"}
	cat, err := m.Discover(context.Background(), ids)
	if err != nil || len(cat.Models) != 3 || !slices.Equal(api.selected, ids[1:2]) || !slices.Equal(cli.selected, []string{ids[0], ids[2]}) {
		t.Fatalf("discovery: %+v %v API=%v CLI=%v", cat, err, api.selected, cli.selected)
	}
	for _, id := range ids {
		effort, endpoint := "medium", api.endpoint
		if ClaudeModel(id) {
			endpoint = cli.endpoint
		}
		q := dataset.Question{Text: "question"}
		mixed, err := Build(m.Endpoint(), q, id, effort, true)
		if err != nil {
			t.Fatal(err)
		}
		direct, err := Build(endpoint, q, id, effort, true)
		if err != nil {
			t.Fatal(err)
		}
		if string(mixed) != string(direct) || EndpointFor(m.Endpoint(), id) != endpoint {
			t.Fatalf("request identity changed for %s", id)
		}
		m.Complete(context.Background(), mixed, func(openrouter.Progress) error { return nil })
	}
	if api.calls != 1 || cli.calls != 2 {
		t.Fatalf("wrong dispatch API=%d CLI=%d", api.calls, cli.calls)
	}
	if _, err := Build(m.Endpoint(), dataset.Question{}, "anthropic/claude-new-model", "medium", true); err == nil {
		t.Fatal("unknown Claude model fell back to OpenRouter")
	}
	if _, err := Build("mixed://v1", dataset.Question{}, ids[1], "medium", true); err == nil {
		t.Fatal("accepted missing routing manifest")
	}
	if !m.Complete(context.Background(), []byte(`{}`), nil).Fatal {
		t.Fatal("accepted missing model")
	}
}

func TestMixedPreflightNeverFallsBackOnClaudeFailure(t *testing.T) {
	api := &routeFixture{endpoint: "https://gateway.test"}
	cli := &routeFixture{endpoint: "claude-cli://v1/2.1.296", fail: true}
	m := Mixed{OpenRouter: api, Claude: cli}
	if _, err := m.Discover(context.Background(), []string{"anthropic/claude-sonnet-5.5", "openai/gpt-6-luna"}); err == nil || len(api.selected) != 0 || api.calls != 0 {
		t.Fatal("Claude failure reached paid backend", err)
	}
}
