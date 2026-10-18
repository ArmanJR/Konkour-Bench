package bench

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

func snapshot() dataset.Snapshot {
	q := dataset.Question{Dataset: "exam", Number: 1, Category: "math", Text: "2+2?", Correct: 4, Options: []dataset.Option{{Number: 1, Text: "1"}, {Number: 2, Text: "2"}, {Number: 3, Text: "3"}, {Number: 4, Text: "4"}}}
	return dataset.Snapshot{Config: dataset.Config{Models: []string{"test/model"}}, Questions: []dataset.Question{q}}
}
func setup(t *testing.T, images bool, efforts string, chat http.HandlerFunc) (*Runner, *storage.Store) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/models":
			modalities := []string{"text"}
			if images {
				modalities = append(modalities, "image")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{map[string]any{"id": "test/model", "architecture": map[string]any{"input_modalities": modalities}, "supported_parameters": []string{"structured_outputs", "reasoning"}, "reasoning": map[string]any{"supported_efforts": json.RawMessage(efforts)}}}})
		case "/models/test/model/endpoints":
			_, _ = w.Write([]byte(`{"data":{"endpoints":[]}}`))
		default:
			chat(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	s, err := storage.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	r := &Runner{Store: s, Client: openrouter.New(srv.URL, "test-key"), Log: slog.New(slog.NewTextHandler(io.Discard, nil)), Sleep: func(context.Context, time.Duration) error { return nil }}
	return r, s
}
func success(w http.ResponseWriter, answer string) {
	_ = json.NewEncoder(w).Encode(map[string]any{"id": "gen-test", "provider": "test-provider", "model": "test/model", "choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"content": answer}}}, "usage": map[string]any{"prompt_tokens": 10, "completion_tokens": 30, "total_tokens": 40, "cost": 0.012, "completion_tokens_details": map[string]any{"reasoning_tokens": 20}}})
}
func TestRunReuseAndAnswerKeyCorrection(t *testing.T) {
	var calls atomic.Int64
	r, s := setup(t, true, `["minimal","medium","max"]`, func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		body, _ := io.ReadAll(req.Body)
		var v map[string]any
		if err := json.Unmarshal(body, &v); err != nil {
			t.Error(err)
		}
		if v["model"] != "test/model:exacto" {
			t.Error("exacto absent")
		}
		success(w, `{"answer":4}`)
	})
	snap := snapshot()
	opts := Options{Concurrency: 2}
	first, err := r.Run(context.Background(), snap, opts)
	if err != nil {
		t.Fatal(err)
	}
	if first.Complete != 3 || first.NewAttempts != 3 || first.NewCost == nil || math.Abs(*first.NewCost-0.036) > 1e-12 {
		t.Fatalf("first: %+v", first)
	}
	snap.Questions[0].Correct = 1
	snap.Questions[0].Category = "changed"
	second, err := r.Run(context.Background(), snap, opts)
	if err != nil {
		t.Fatal(err)
	}
	if second.Cached != 3 || second.NewAttempts != 0 || calls.Load() != 3 || second.NewCost == nil || *second.NewCost != 0 {
		t.Fatalf("second: %+v calls %d", second, calls.Load())
	}
	jobs, err := s.Jobs(second.RunID, Protocol())
	if err != nil || len(jobs) != 3 {
		t.Fatalf("jobs: %v %v", jobs, err)
	}
	snap.Questions[0].Text = "changed question"
	third, err := r.Run(context.Background(), snap, opts)
	if err != nil {
		t.Fatal(err)
	}
	if third.NewAttempts != 3 || calls.Load() != 6 {
		t.Fatalf("changed question not sent: %+v", third)
	}
}
func TestCompatibilityAndImageSkip(t *testing.T) {
	var calls atomic.Int64
	r, _ := setup(t, false, `["low","high"]`, func(w http.ResponseWriter, req *http.Request) { calls.Add(1); success(w, `{"answer":4}`) })
	snap := snapshot()
	q := snap.Questions[0]
	q.Number = 2
	q.Options = append([]dataset.Option(nil), q.Options...)
	q.Options[0].Asset = &dataset.Asset{MIME: "image/png", Bytes: []byte{1}}
	snap.Questions = append(snap.Questions, q)
	s, err := r.Run(context.Background(), snap, Options{Concurrency: 2})
	if err != nil {
		t.Fatal(err)
	}
	if s.Complete != 2 || s.Unsupported != 1 || s.Skipped != 3 || calls.Load() != 2 {
		t.Fatalf("summary %+v calls %d", s, calls.Load())
	}
}
func TestExplicitEffortFallbackAndTransientRetry(t *testing.T) {
	var calls atomic.Int64
	var lowCalls atomic.Int64
	r, _ := setup(t, true, `null`, func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		var body openrouter.Request
		_ = json.NewDecoder(req.Body).Decode(&body)
		if body.Reasoning.Effort == "minimal" {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(`{"error":{"message":"unsupported reasoning effort minimal"}}`))
			return
		}
		if body.Reasoning.Effort != "low" {
			t.Errorf("unexpected fallback %s", body.Reasoning.Effort)
		}
		if lowCalls.Add(1) == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(429)
			return
		}
		success(w, `{"answer":4}`)
	})
	var delays []time.Duration
	r.Sleep = func(_ context.Context, d time.Duration) error { delays = append(delays, d); return nil }
	opts := Options{Concurrency: 1, Selection: storage.Selection{Efforts: []string{"minimal"}}}
	s, err := r.Run(context.Background(), snapshot(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if s.Complete != 1 || calls.Load() != 3 || len(delays) != 1 || delays[0] != 2*time.Second {
		t.Fatalf("summary=%+v calls=%d delays=%v", s, calls.Load(), delays)
	}
	s, err = r.Run(context.Background(), snapshot(), opts)
	if err != nil || s.Cached != 1 || calls.Load() != 3 {
		t.Fatalf("fallback cache %+v %v calls=%d", s, err, calls.Load())
	}
}
func TestInvalidAnswerIsNotResampled(t *testing.T) {
	var calls atomic.Int64
	r, s := setup(t, true, `["medium"]`, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); success(w, "I choose option 4") })
	opts := Options{Concurrency: 1, Selection: storage.Selection{Efforts: []string{"medium"}}}
	first, err := r.Run(context.Background(), snapshot(), opts)
	if err != nil {
		t.Fatal(err)
	}
	jobs, _ := s.Jobs(first.RunID, Protocol())
	req, err := s.Request(jobs[0].Hash)
	if err != nil || req.Result.Invalid == "" || req.Result.Answer != 0 {
		t.Fatalf("invalid result %+v %v", req, err)
	}
	second, err := r.Run(context.Background(), snapshot(), opts)
	if err != nil || second.Cached != 1 || calls.Load() != 1 {
		t.Fatalf("invalid answer resampled: %+v %v", second, err)
	}
}
func TestUnrelated400DoesNotChangeEffort(t *testing.T) {
	var calls atomic.Int64
	r, _ := setup(t, true, `null`, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error":"invalid schema"}`))
	})
	s, err := r.Run(context.Background(), snapshot(), Options{Concurrency: 1, Selection: storage.Selection{Efforts: []string{"max"}}})
	if err == nil || s.Failed != 1 || calls.Load() != 1 {
		t.Fatalf("summary %+v err=%v calls=%d", s, err, calls.Load())
	}
}
func TestUnknownResponseNotAutomaticallyResent(t *testing.T) {
	var calls atomic.Int64
	r, _ := setup(t, true, `["medium"]`, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); _, _ = w.Write([]byte(`{"broken":`)) })
	opts := Options{Concurrency: 1, Selection: storage.Selection{Efforts: []string{"medium"}}}
	for i := 0; i < 2; i++ {
		s, err := r.Run(context.Background(), snapshot(), opts)
		if err == nil || s.Failed != 1 {
			t.Fatalf("summary %+v %v", s, err)
		}
	}
	if calls.Load() != 1 {
		t.Fatal("uncertain generation automatically resent")
	}
}
func TestHashIncludesImageBytesAndParameters(t *testing.T) {
	q := snapshot().Questions[0]
	q.Asset = &dataset.Asset{MIME: "image/png", Bytes: []byte{1}}
	a, _ := openrouter.Build(q, "test/model", "medium")
	b, _ := openrouter.Build(q, "test/model", "max")
	q.Asset.Bytes = []byte{2}
	c, _ := openrouter.Build(q, "test/model", "medium")
	if RequestHash("http://test", a) == RequestHash("http://test", b) || RequestHash("http://test", a) == RequestHash("http://test", c) || RequestHash("http://a", a) == RequestHash("http://b", a) {
		t.Fatal("request hash omitted significant input")
	}
}

func TestReconcileMissingAccounting(t *testing.T) {
	var generations atomic.Int64
	r, s := setup(t, true, `["medium"]`, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/generation" {
			if req.URL.Query().Get("id") != "gen-test" {
				t.Error("wrong generation ID")
			}
			_, _ = w.Write([]byte(`{"data":{"total_cost":0.02,"native_tokens_prompt":10,"native_tokens_completion":30,"native_tokens_reasoning":20,"provider_name":"test"}}`))
			return
		}
		generations.Add(1)
		_, _ = w.Write([]byte(`{"id":"gen-test","choices":[{"finish_reason":"stop","message":{"content":"{\"answer\":4}"}}]}`))
	})
	summary, err := r.Run(context.Background(), snapshot(), Options{Concurrency: 1, Selection: storage.Selection{Efforts: []string{"medium"}}})
	if err != nil {
		t.Fatal(err)
	}
	if summary.NewCost != nil || summary.UnknownCostAttempts != 1 {
		t.Fatalf("missing cost treated as zero: %+v", summary)
	}
	n, err := Reconcile(context.Background(), s, r.Client)
	if err != nil || n != 1 {
		t.Fatalf("reconcile %d %v", n, err)
	}
	jobs, _ := s.Jobs(summary.RunID, Protocol())
	req, err := s.Request(jobs[0].Hash)
	if err != nil {
		t.Fatal(err)
	}
	u := req.Result.Generation.Usage
	if u.Total == nil || *u.Total != 40 || u.Cost == nil || *u.Cost != 0.02 || *u.CompletionDetails.Reasoning != 20 || generations.Load() != 1 {
		t.Fatalf("reconciliation usage %+v generations %d", u, generations.Load())
	}
	n, err = Reconcile(context.Background(), s, r.Client)
	if err != nil || n != 0 {
		t.Fatalf("reconciled twice: %d %v", n, err)
	}
}

func TestAuthenticationStopsRunAndServerErrorsAreBounded(t *testing.T) {
	for _, status := range []int{401, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int64
			r, _ := setup(t, true, `["medium"]`, func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"error":"test failure"}`))
			})
			s, err := r.Run(context.Background(), snapshot(), Options{Concurrency: 1, Selection: storage.Selection{Efforts: []string{"medium"}}})
			want := int64(1)
			if status == 500 {
				want = 3
			}
			if err == nil || s.Failed != 1 || calls.Load() != want {
				t.Fatalf("status=%d summary=%+v calls=%d err=%v", status, s, calls.Load(), err)
			}
		})
	}
}
