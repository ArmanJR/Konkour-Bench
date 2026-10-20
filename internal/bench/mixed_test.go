package bench

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"konkour-bench/internal/backbone"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

type mixedBackendFixture struct {
	endpoint            string
	calls, active, peak atomic.Int32
	fatal               bool
	unknown             bool
}

func (c *mixedBackendFixture) Endpoint() string { return c.endpoint }
func (c *mixedBackendFixture) Discover(_ context.Context, ids []string) (openrouter.Catalog, error) {
	cat := openrouter.Catalog{Models: map[string]openrouter.Model{}, Raw: json.RawMessage(`{}`)}
	for _, id := range ids {
		var model openrouter.Model
		b, _ := json.Marshal(map[string]any{"id": id, "architecture": map[string]any{"input_modalities": []string{"text", "image"}}, "supported_parameters": []string{"structured_outputs"}, "reasoning": map[string]any{"supported_efforts": []string{"low", "medium", "max"}}})
		if err := json.Unmarshal(b, &model); err != nil {
			return cat, err
		}
		cat.Models[id] = model
	}
	return cat, nil
}
func (c *mixedBackendFixture) Complete(ctx context.Context, body []byte, save func(openrouter.Progress) error) openrouter.HTTPResult {
	call := c.calls.Add(1)
	active := c.active.Add(1)
	defer c.active.Add(-1)
	for {
		peak := c.peak.Load()
		if active <= peak || c.peak.CompareAndSwap(peak, active) {
			break
		}
	}
	if err := Wait(ctx, 10*time.Millisecond); err != nil {
		return openrouter.HTTPResult{Error: err.Error()}
	}
	if c.fatal {
		return openrouter.HTTPResult{Error: "backend account limit", Fatal: true}
	}
	var req struct{ Model string }
	_ = json.Unmarshal(body, &req)
	id := fmt.Sprintf("gen-mixed-%d", call)
	if err := save(openrouter.Progress{GenerationID: id, Raw: `{"event":"test"}`}); err != nil {
		return openrouter.HTTPResult{Error: err.Error()}
	}
	if c.unknown && call == 1 {
		return openrouter.HTTPResult{Error: "interrupted stream", GenerationID: id}
	}
	var cost any = 0.02
	if backbone.IsClaude(c.endpoint) {
		cost = nil
	}
	b, _ := json.Marshal(map[string]any{"id": id, "model": req.Model, "usage": map[string]any{"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30, "cost": cost}, "choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]string{"content": `{"answer":4}`}}}})
	return openrouter.HTTPResult{Status: 200, GenerationID: id, Body: b}
}
func mixedSetup(t *testing.T) (*Runner, *mixedBackendFixture, *mixedBackendFixture, dataset.Snapshot) {
	t.Helper()
	s, err := storage.Open(filepath.Join(t.TempDir(), "bench.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	api := &mixedBackendFixture{endpoint: "https://gateway.test"}
	cli := &mixedBackendFixture{endpoint: "claude-cli://v1/2.1.296"}
	snap := snapshot()
	snap.Config.Models = []string{"test/api", "anthropic/claude-sonnet-5.5", "anthropic/claude-opus-5.5"}
	q := snap.Questions[0]
	q.Number = 2
	q.Text = "another question"
	snap.Questions = append(snap.Questions, q)
	r := &Runner{Store: s, Client: openrouter.New(api.endpoint, ""), Backend: &backbone.Mixed{OpenRouter: api, Claude: cli}, Log: slog.New(slog.NewTextHandler(io.Discard, nil)), Sleep: func(context.Context, time.Duration) error { return nil }}
	return r, api, cli, snap
}

func TestMixedRunCacheBudgetAndConcurrency(t *testing.T) {
	r, api, cli, snap := mixedSetup(t)
	opts := Options{Concurrency: 4, BudgetUSD: 1, Selection: storage.Selection{Efforts: []string{"medium"}}}
	// Seed an existing direct OpenRouter result: switching routing must reuse it.
	r.Backend = api
	seed := opts
	seed.Selection.Models = []string{"test/api"}
	seed.Selection.Questions = []string{"exam:1"}
	if _, err := r.Run(context.Background(), snap, seed); err != nil {
		t.Fatal(err)
	}
	r.Backend = &backbone.Mixed{OpenRouter: api, Claude: cli}
	first, err := r.Run(context.Background(), snap, opts)
	if err != nil || first.Complete != 6 || first.Cached != 1 || first.SubscriptionAttempts != 4 || first.KnownCost != 0.02 || first.UnknownCostAttempts != 0 {
		t.Fatalf("mixed summary: %+v %v", first, err)
	}
	if api.calls.Load() != 2 || cli.calls.Load() != 4 || cli.peak.Load() != 1 {
		t.Fatalf("calls/concurrency: API=%d CLI=%d peak=%d", api.calls.Load(), cli.calls.Load(), cli.peak.Load())
	}
	second, err := r.Run(context.Background(), snap, opts)
	if err != nil || second.Cached != 6 || second.NewAttempts != 0 {
		t.Fatalf("mixed cache: %+v %v", second, err)
	}
	attempts, err := r.Store.Attempts("", first.RunID)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range attempts {
		if a.Details.Backend != "claude" && a.Details.Backend != "openrouter" {
			t.Fatal("missing attempt backend", a.ID)
		}
	}
}

func TestMixedBackendFailureDoesNotCancelOtherBackend(t *testing.T) {
	for _, failing := range []string{"claude", "openrouter"} {
		t.Run(failing, func(t *testing.T) {
			r, api, cli, snap := mixedSetup(t)
			if failing == "claude" {
				cli.fatal = true
			} else {
				api.fatal = true
			}
			result, err := r.Run(context.Background(), snap, Options{Concurrency: 4, Selection: storage.Selection{Efforts: []string{"medium"}}})
			if err == nil {
				t.Fatal("fatal failure not reported")
			}
			want := 2
			if failing == "openrouter" {
				want = 4
			}
			if result.Complete != want {
				t.Fatalf("other backend interrupted: %+v %v", result, err)
			}
		})
	}
}

func TestMixedBudgetStopsAPIWithoutBlockingSubscription(t *testing.T) {
	r, api, cli, snap := mixedSetup(t)
	result, err := r.Run(context.Background(), snap, Options{Concurrency: 4, BudgetUSD: 0.01, Selection: storage.Selection{Efforts: []string{"medium"}}})
	if err == nil || api.calls.Load() != 1 || cli.calls.Load() != 4 || result.SubscriptionAttempts != 4 {
		t.Fatalf("budget routing: %+v %v API=%d CLI=%d", result, err, api.calls.Load(), cli.calls.Load())
	}
}

func TestMixedReconciliationUsesOnlyAPIRoute(t *testing.T) {
	r, api, _, snap := mixedSetup(t)
	var lookups atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		lookups.Add(1)
		if req.URL.Path != "/generation" || req.URL.Query().Get("id") != "gen-mixed-1" {
			t.Error("wrong accounting route", req.URL)
		}
		fmt.Fprint(w, `{"data":{"id":"gen-mixed-1","total_cost":0.01,"native_tokens_prompt":10,"native_tokens_completion":20,"finish_reason":"cancelled"}}`)
	}))
	defer srv.Close()
	api.endpoint = srv.URL
	api.unknown = true
	r.Client = openrouter.New(srv.URL, "")
	_, err := r.Run(context.Background(), snap, Options{Concurrency: 4, BudgetUSD: 1, Selection: storage.Selection{Efforts: []string{"medium"}}})
	if err == nil || lookups.Load() != 1 {
		t.Fatalf("accounting: lookups=%d err=%v", lookups.Load(), err)
	}
	if n, err := Reconcile(context.Background(), r.Store, r.Client); err != nil || n != 0 || lookups.Load() != 1 {
		t.Fatalf("CLI session reconciled as API: %d %v", n, err)
	}
}
