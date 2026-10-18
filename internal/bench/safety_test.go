package bench

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

func TestDeadlineKeepsIDAndNeverRetries(t *testing.T) {
	var calls atomic.Int64
	r, s := setup(t, true, `["medium"]`, func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("X-Generation-Id", "gen-timeout")
		w.(http.Flusher).Flush()
		<-req.Context().Done()
	})
	opts := Options{Concurrency: 1, Timeout: 100 * time.Millisecond, Selection: storage.Selection{Efforts: []string{"medium"}}}
	first, err := r.Run(context.Background(), snapshot(), opts)
	if err == nil || first.Failed != 1 {
		t.Fatalf("timeout %+v %v", first, err)
	}
	attempts, err := s.Attempts("", first.RunID)
	if err != nil || len(attempts) != 1 || attempts[0].Details.GenerationID != "gen-timeout" {
		t.Fatalf("attempts %+v %v", attempts, err)
	}
	req, _ := s.Request(attempts[0].Hash)
	if req.State != "unknown" || !strings.Contains(req.Note, "deadline_exceeded") || req.Result.Answer != 0 {
		t.Fatal(req)
	}
	_, err = r.Run(context.Background(), snapshot(), opts)
	if err == nil || calls.Load() != 1 {
		t.Fatalf("timeout automatically retried: %d %v", calls.Load(), err)
	}
}

func TestStreamingRequestStillHasNoTokenCaps(t *testing.T) {
	r, _ := setup(t, true, `["max"]`, func(w http.ResponseWriter, req *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["stream"] != true || body["model"] != "test/model:exacto" {
			t.Error(body)
		}
		for _, k := range []string{"max_tokens", "max_completion_tokens"} {
			if _, ok := body[k]; ok {
				t.Error("token cap", k)
			}
		}
		if _, ok := body["reasoning"].(map[string]any)["max_tokens"]; ok {
			t.Error("reasoning cap")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"id\":\"gen-sse\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"{\\\"answer\\\":4}\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n")
	})
	summary, err := r.Run(context.Background(), snapshot(), Options{Concurrency: 1, Selection: storage.Selection{Efforts: []string{"max"}}})
	if err != nil || summary.Complete != 1 {
		t.Fatalf("%+v %v", summary, err)
	}
}

func TestLegacyCacheAndUncertainTransportTransition(t *testing.T) {
	for _, state := range []string{"complete", "unknown"} {
		t.Run(state, func(t *testing.T) {
			var calls atomic.Int64
			r, s := setup(t, true, `["medium"]`, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); success(w, `{"answer":4}`) })
			snap := snapshot()
			body, _ := openrouter.Build(snap.Questions[0], "test/model", "medium")
			oldRun := storage.Run{ID: "legacy", Protocol: Protocol(), BaseURL: r.Client.BaseURL}
			if err := s.AddRun(oldRun); err != nil {
				t.Fatal(err)
			}
			old, _, err := s.Claim(oldRun.ID, RequestHash(r.Client.BaseURL, body), body, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Finish(old, state, openrouter.HTTPResult{}, storage.Result{Answer: 4}, "legacy"); err != nil {
				t.Fatal(err)
			}
			opts := Options{Concurrency: 1, Selection: storage.Selection{Efforts: []string{"medium"}}}
			summary, err := r.Run(context.Background(), snap, opts)
			if calls.Load() != 0 {
				t.Fatal("legacy request resent")
			}
			if state == "complete" {
				if err != nil || summary.Cached != 1 {
					t.Fatal(summary, err)
				}
				return
			}
			if err == nil {
				t.Fatal("unknown treated as complete")
			}
			opts.RetryUnknown = true
			if _, err := r.Run(context.Background(), snap, opts); err != nil {
				t.Fatal(err)
			}
			opts.RetryUnknown = false
			summary, err = r.Run(context.Background(), snap, opts)
			if err != nil || summary.Cached != 1 || calls.Load() != 1 {
				t.Fatalf("stream cache blocked by old unknown: %+v %v calls=%d", summary, err, calls.Load())
			}
		})
	}
}

func TestBudgetStopsDispatchAfterCostOrUnknown(t *testing.T) {
	for _, unknown := range []bool{false, true} {
		t.Run(fmt.Sprint(unknown), func(t *testing.T) {
			var calls atomic.Int64
			r, _ := setup(t, true, `["medium"]`, func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				if unknown {
					fmt.Fprint(w, `{"id":"gen-no-cost","choices":[{"finish_reason":"stop","message":{"content":"{\"answer\":4}"}}]}`)
				} else {
					success(w, `{"answer":4}`)
				}
			})
			snap := snapshot()
			q := snap.Questions[0]
			q.Number = 2
			q.Text = "another question"
			snap.Questions = append(snap.Questions, q)
			result, err := r.Run(context.Background(), snap, Options{Concurrency: 1, BudgetUSD: 0.01, Selection: storage.Selection{Efforts: []string{"medium"}}})
			if !errors.Is(err, ErrBudget) || calls.Load() != 1 || result.Complete != 1 {
				t.Fatalf("budget %+v %v calls %d", result, err, calls.Load())
			}
		})
	}
}

func TestCancelledReconciliationDoesNotInventAnswer(t *testing.T) {
	var calls atomic.Int64
	r, s := setup(t, true, `["medium"]`, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/generation" {
			fmt.Fprint(w, `{"data":{"id":"gen-cancelled","cancelled":true,"usage":0.11,"native_tokens_prompt":378,"native_tokens_completion":61102,"native_tokens_reasoning":61102,"provider_name":"Wafer"}}`)
			return
		}
		calls.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"id\":\"gen-cancelled\",\"choices\":[{\"delta\":{\"reasoning\":\"thinking\"}}]}\n\n")
	})
	run, err := r.Run(context.Background(), snapshot(), Options{Concurrency: 1, Selection: storage.Selection{Efforts: []string{"medium"}}})
	if err == nil {
		t.Fatal("truncated stream succeeded")
	}
	n, err := Reconcile(context.Background(), s, r.Client)
	if err != nil || n != 1 {
		t.Fatal(n, err)
	}
	attempts, _ := s.Attempts("", run.RunID)
	req, _ := s.Request(attempts[0].Hash)
	if req.State != "unknown" || req.Result.Answer != 0 || *req.Result.Generation.Usage.Cost != 0.11 || *req.Result.Generation.Usage.Total != 61480 || calls.Load() != 1 {
		t.Fatalf("reconciliation %+v calls=%d", req, calls.Load())
	}
	if *attempts[0].Usage().Cost != 0.11 {
		t.Fatal("reconciled charge absent from attempt ledger")
	}
}

func TestManualReconcileRejectsMismatchedIdentity(t *testing.T) {
	var data map[string]any
	r, s := setup(t, true, `["medium"]`, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/generation" {
			json.NewEncoder(w).Encode(map[string]any{"data": data})
			return
		}
		fmt.Fprint(w, `{broken}`)
	})
	run, _ := r.Run(context.Background(), snapshot(), Options{Concurrency: 1, Selection: storage.Selection{Efforts: []string{"medium"}}})
	attempts, _ := s.Attempts("", run.RunID)
	a := attempts[0]
	base := func() map[string]any {
		return map[string]any{"id": "gen-manual", "model": "test/model-20261018", "created_at": a.Started, "cancelled": true, "usage": 0.1}
	}
	for _, field := range []string{"id", "model", "created_at"} {
		data = base()
		data[field] = "mismatch"
		if err := ReconcileID(context.Background(), s, r.Client, a.ID, "gen-manual"); err == nil {
			t.Error("accepted wrong", field)
		}
	}
	data = base()
	if err := ReconcileID(context.Background(), s, r.Client, a.ID, "gen-manual"); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Attempts("", run.RunID)
	if got[0].Details.GenerationID != "gen-manual" || *got[0].Usage().Cost != 0.1 {
		t.Fatal(got)
	}
}
