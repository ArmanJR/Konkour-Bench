package report

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"konkour-bench/internal/backbone"
	"konkour-bench/internal/bench"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

type fixtureBackend struct {
	endpoint string
	calls    int
	fatal    bool
	paid     bool
}

func (c *fixtureBackend) Endpoint() string { return c.endpoint }
func (c *fixtureBackend) Discover(_ context.Context, ids []string) (openrouter.Catalog, error) {
	var m openrouter.Model
	err := json.Unmarshal([]byte(`{"id":"anthropic/claude-sonnet-5.5","architecture":{"input_modalities":["text","image"]},"supported_parameters":["structured_outputs"],"reasoning":{"supported_efforts":["low","medium","max"]}}`), &m)
	cat := openrouter.Catalog{Models: map[string]openrouter.Model{}}
	for _, id := range ids {
		m.ID = id
		cat.Models[id] = m
	}
	return cat, err
}
func (c *fixtureBackend) Complete(_ context.Context, body []byte, save func(openrouter.Progress) error) openrouter.HTTPResult {
	c.calls++
	if c.fatal {
		return openrouter.HTTPResult{Error: "subscription rate limit", Fatal: true}
	}
	if err := save(openrouter.Progress{GenerationID: "cli-session", Provider: "Claude CLI (subscription)", Raw: `{"type":"result","structured_output":{"answer":2}}`}); err != nil {
		return openrouter.HTTPResult{Error: err.Error()}
	}
	response := `{"id":"cli-session","provider":"Claude CLI (subscription)","model":"claude-sonnet-5-5","usage":{"prompt_tokens":10,"completion_tokens":20,"total_tokens":30,"cost":null,"estimated_api_cost_usd":0.1},"choices":[{"finish_reason":"stop","message":{"content":"{\"answer\":2}"}}]}`
	if c.paid {
		response = strings.Replace(response, `"cost":null`, `"cost":0.1`, 1)
	}
	return openrouter.HTTPResult{Status: 200, GenerationID: "cli-session", Body: []byte(response)}
}

func TestMixedReportKeepsSubscriptionCostOutOfPaidRankings(t *testing.T) {
	s, err := storage.Open(filepath.Join(t.TempDir(), "mixed.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	api := &fixtureBackend{endpoint: "https://gateway.test", paid: true}
	cli := &fixtureBackend{endpoint: "claude-cli://v1/2.1.296"}
	mixed := &backbone.Mixed{OpenRouter: api, Claude: cli}
	r := bench.Runner{Store: s, Backend: mixed, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	snap := dataset.Snapshot{Config: dataset.Config{Models: []string{"test/api", "anthropic/claude-sonnet-5.5"}}, Questions: []dataset.Question{{Dataset: "exam", Number: 1, Category: "math", Text: "test", Correct: 2}}}
	result, err := r.Run(context.Background(), snap, bench.Options{Concurrency: 2, BudgetUSD: 1, Selection: storage.Selection{Efforts: []string{"medium"}}})
	if err != nil {
		t.Fatal(err)
	}
	saved, err := s.Run(result.RunID)
	if err != nil {
		t.Fatal(err)
	}
	for _, scope := range []struct{ id, endpoint string }{{saved.ID, saved.BaseURL}, {"", "mixed://latest"}} {
		doc, err := Build(s, snap, scope.id, scope.endpoint, saved.Selection)
		if err != nil {
			t.Fatal(err)
		}
		if doc.Backend != "mixed" || len(doc.Outcomes) != 2 || doc.Spending[0].KnownCost != 0.1 || doc.Spending[0].UnknownCost != 0 || doc.Spending[0].SubscriptionAttempts != 1 {
			t.Fatalf("wrong mixed report: %+v", doc)
		}
		for _, o := range doc.Outcomes {
			if !o.Correct || o.Backend != backbone.Name(backbone.EndpointFor(saved.BaseURL, o.Model)) {
				t.Fatalf("wrong outcome: %+v", o)
			}
		}
		for _, row := range doc.Rows {
			if row.Accuracy == nil || *row.Accuracy != 1 {
				t.Fatalf("missing mixed score: %+v", row)
			}
			if backbone.ClaudeModel(row.Model) && (row.Cost != nil || row.CostFrontier) {
				t.Fatal("subscription ranked as paid API cost")
			}
			if row.Model == "test/api" && (row.Cost == nil || *row.Cost != 0.1 || !row.CostFrontier) {
				t.Fatalf("API cost excluded: %+v", row)
			}
		}
	}
}

func TestClaudeRunnerCacheReportingAndAccountingIsolation(t *testing.T) {
	s, err := storage.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	q := dataset.Question{Number: 1, Dataset: "exam", Category: "math", Text: "2+2?", Correct: 2, Options: []dataset.Option{{Number: 1, Text: "3"}, {Number: 2, Text: "4"}, {Number: 3, Text: "5"}, {Number: 4, Text: "6"}}}
	// Inline image bytes are enough here: dataset decoding has separate tests.
	q.Options[0].Asset = &dataset.Asset{MIME: "image/png", Bytes: []byte{1, 2, 3}}
	snap := dataset.Snapshot{Config: dataset.Config{Models: []string{"anthropic/claude-sonnet-5.5"}}, Questions: []dataset.Question{q}}
	c := &fixtureBackend{endpoint: "claude-cli://v1/2.1.296"}
	r := bench.Runner{Store: s, Backend: c, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	opts := bench.Options{Concurrency: 1}
	first, err := r.Run(context.Background(), snap, opts)
	if err != nil {
		t.Fatal(err)
	}
	if first.Complete != 3 || first.Skipped != 0 || first.NewCost != nil || first.KnownCost != 0 || first.UnknownCostAttempts != 3 || first.EstimatedCost < 0.299 {
		t.Fatalf("summary: %+v", first)
	}
	saved, err := s.Run(first.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Protocol == bench.Protocol() || saved.BaseURL != c.endpoint {
		t.Fatalf("lost backend identity: %+v", saved)
	}
	jobs, err := s.Jobs(first.RunID, saved.Protocol)
	if err != nil {
		t.Fatal(jobs, err)
	}
	completed := 0
	for _, j := range jobs {
		if j.Status != "complete" {
			continue
		}
		completed++
		req, err := s.Request(j.Hash)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(req.Body), `"--json-schema"`) || strings.Contains(string(req.Body), "max_tokens") {
			t.Fatalf("incorrect persisted CLI request: %s", req.Body)
		}
		if j.Tier == "minimal" && j.Effort != "low" {
			t.Fatal("wrong minimal mapping", j)
		}
	}
	if completed != 3 {
		t.Fatalf("completed jobs: %d", completed)
	}
	doc, err := Build(s, snap, first.RunID, c.endpoint, saved.Selection)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Backend != "claude" || !strings.Contains(doc.CostBasis, "subscription") || len(doc.Outcomes) != 3 {
		t.Fatalf("report identity: %+v", doc)
	}
	for _, row := range doc.Rows {
		if row.Cost != nil || row.CostFrontier || row.Pareto || row.Accuracy == nil || *row.Accuracy != 1 || row.Tokens == nil || !row.TokenFrontier {
			t.Fatalf("incorrect CLI rankings: %+v", row)
		}
	}
	if err := Write(filepath.Join(t.TempDir(), "report"), doc); err != nil {
		t.Fatal(err)
	}
	// Corrected grading metadata reuses answers and changes only scoring.
	snap.Questions[0].Correct = 1
	second, err := r.Run(context.Background(), snap, opts)
	if err != nil || second.Cached != 3 || c.calls != 3 {
		t.Fatalf("cache: %+v %v calls=%d", second, err, c.calls)
	}
	doc, err = Build(s, snap, "", c.endpoint, saved.Selection)
	if err != nil || doc.Outcomes[0].Correct {
		t.Fatalf("rescoring: %v %v", doc.Outcomes, err)
	}
	// Reconciliation cannot send CLI session IDs or private data to OpenRouter.
	count, err := bench.Reconcile(context.Background(), s, openrouter.New("http://127.0.0.1:1", ""))
	if err != nil || count != 0 {
		t.Fatalf("CLI reconciliation not skipped: %d %v", count, err)
	}
	// The same model/prompt from another transport must invoke that transport.
	c.endpoint = "https://other-backend.test"
	third, err := r.Run(context.Background(), snap, opts)
	if err != nil || third.Cached != 0 || c.calls != 6 {
		t.Fatalf("backends mixed: %+v %v", third, err)
	}
	// A CLI upgrade also has its own namespace, including report cache checks.
	c.endpoint = "claude-cli://v1/2.1.297"
	fourth, err := r.Run(context.Background(), snap, opts)
	if err != nil || fourth.Cached != 0 || c.calls != 9 {
		t.Fatalf("version cache: %+v %v", fourth, err)
	}
	// Current-dataset reports are offline, use the last CLI version, and do
	// not accidentally include the non-Claude models in the default config.
	snap.Config.Models = append(snap.Config.Models, "openai/gpt-6.1-sol")
	doc, err = Build(s, snap, "", "claude-cli://latest", storage.Selection{})
	if err != nil || len(doc.Models) != 1 || len(doc.Outcomes) != 3 || doc.Outcomes[0].RunID != fourth.RunID || doc.Outcomes[0].Correct {
		t.Fatalf("current CLI report: %+v %v", doc, err)
	}
	for _, spend := range doc.Spending {
		if spend.RunID == third.RunID {
			t.Fatal("OpenRouter charges leaked into the Claude ledger")
		}
	}
}

func TestClaudeFatalFailureStopsDispatchAndRequiresExplicitRetry(t *testing.T) {
	s, err := storage.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := &fixtureBackend{endpoint: "claude-cli://v1/2.1.296", fatal: true}
	r := bench.Runner{Store: s, Backend: c, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	snap := dataset.Snapshot{Config: dataset.Config{Models: []string{"anthropic/claude-sonnet-5.5"}}, Questions: []dataset.Question{{Dataset: "exam", Number: 1, Text: "test"}}}
	if _, err := r.Run(context.Background(), snap, bench.Options{Concurrency: 1, BudgetUSD: 10}); err == nil || c.calls != 0 {
		t.Fatal("subscription accepted a fake monetary cap", err)
	}
	opts := bench.Options{Concurrency: 1, Selection: storage.Selection{Efforts: []string{"medium", "max"}}}
	first, err := r.Run(context.Background(), snap, opts)
	if err == nil || first.Failed != 1 || c.calls != 1 {
		t.Fatalf("fatal: %+v %v", first, err)
	}
	c.fatal = false
	opts.Selection.Efforts = []string{"medium"}
	_, err = r.Run(context.Background(), snap, opts)
	if err == nil || c.calls != 1 {
		t.Fatal("uncertain CLI request was automatically repeated")
	}
	opts.RetryUnknown = true
	last, err := r.Run(context.Background(), snap, opts)
	if err != nil || last.Complete != 1 || c.calls != 2 {
		t.Fatalf("explicit retry: %+v %v", last, err)
	}
}
