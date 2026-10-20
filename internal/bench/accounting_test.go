package bench

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

func TestTimeoutReconcilesAndContinuesWithoutRetry(t *testing.T) {
	for _, limit := range []float64{1, 0.04} {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			var calls, metadata atomic.Int64
			r, store := setup(t, true, `["medium"]`, func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/generation" {
					if req.URL.Query().Get("id") != "gen-timeout" {
						t.Error("incorrect generation requested")
					}
					if metadata.Add(1) == 1 {
						http.Error(w, "accounting not ready", http.StatusNotFound)
						return
					}
					fmt.Fprint(w, `{"data":{"id":"gen-timeout","cancelled":true,"total_cost":0.05,"native_tokens_prompt":10,"native_tokens_completion":100}}`)
					return
				}
				if calls.Add(1) == 1 {
					w.Header().Set("Content-Type", "text/event-stream")
					w.Header().Set("X-Generation-Id", "gen-timeout")
					w.(http.Flusher).Flush()
					<-req.Context().Done()
					return
				}
				if metadata.Load() < 2 {
					t.Error("dispatched before accounting was resolved")
				}
				success(w, `{"answer":4}`)
			})
			snap := snapshot()
			q := snap.Questions[0]
			q.Number, q.Text = 2, "another question"
			snap.Questions = append(snap.Questions, q)
			opts := Options{Concurrency: 1, Timeout: 100 * time.Millisecond, BudgetUSD: limit, Selection: storage.Selection{Efforts: []string{"medium"}}}
			result, err := r.Run(context.Background(), snap, opts)
			wantCalls, wantComplete, wantCost := int64(2), 1, 0.062
			if limit < 0.05 {
				wantCalls, wantComplete, wantCost = 1, 0, 0.05
				if !errors.Is(err, ErrBudget) {
					t.Fatalf("reconciled charge did not enforce budget: %v", err)
				}
			} else if err == nil || errors.Is(err, ErrBudget) {
				t.Fatalf("expected only the unresolved timeout in final summary: %v", err)
			}
			if calls.Load() != wantCalls || metadata.Load() != 2 || result.Complete != wantComplete || result.Failed != 1 || result.UnknownCostAttempts != 0 || result.NewCost == nil || math.Abs(*result.NewCost-wantCost) > 1e-9 {
				t.Fatalf("result=%+v calls=%d metadata=%d err=%v", result, calls.Load(), metadata.Load(), err)
			}
			attempts, err := store.Attempts("", result.RunID)
			if err != nil || len(attempts) != int(wantCalls) {
				t.Fatalf("attempts=%v err=%v", attempts, err)
			}
			request, err := store.Request(attempts[0].Hash)
			if err != nil || request.State != "unknown" || request.Result.Answer != 0 || !strings.Contains(request.Note, "deadline_exceeded") {
				t.Fatalf("timeout outcome changed: %+v %v", request, err)
			}
			// Accounting recovery must not make an interrupted answer retryable.
			if limit > 0.05 {
				again, err := r.Run(context.Background(), snap, opts)
				if err == nil || again.Cached != 1 || again.NewAttempts != 0 || calls.Load() != 2 {
					t.Fatalf("interrupted request resent: %+v %v calls=%d", again, err, calls.Load())
				}
			}
		})
	}
}

func TestFinalAnswerMissingCostReconcilesAutomatically(t *testing.T) {
	var calls, metadata atomic.Int64
	r, _ := setup(t, true, `["medium"]`, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/generation" {
			if metadata.Add(1) == 1 {
				fmt.Fprint(w, `{"data":{"id":"gen-missing-cost","total_cost":0.02}}`)
				return
			}
			fmt.Fprint(w, `{"data":{"id":"gen-missing-cost","finish_reason":"stop","total_cost":0.02,"native_tokens_prompt":10,"native_tokens_completion":30}}`)
			return
		}
		calls.Add(1)
		fmt.Fprint(w, `{"id":"gen-missing-cost","choices":[{"finish_reason":"stop","message":{"content":"{\"answer\":4}"}}]}`)
	})
	result, err := r.Run(context.Background(), snapshot(), Options{Concurrency: 1, BudgetUSD: 1, Selection: storage.Selection{Efforts: []string{"medium"}}})
	if err != nil || result.Complete != 1 || result.NewCost == nil || *result.NewCost != 0.02 || calls.Load() != 1 || metadata.Load() != 2 {
		t.Fatalf("result=%+v calls=%d metadata=%d err=%v", result, calls.Load(), metadata.Load(), err)
	}
}

func TestUnrecoverableAccountingStopsWithoutMoreInference(t *testing.T) {
	for _, missingID := range []bool{false, true} {
		t.Run(fmt.Sprint(missingID), func(t *testing.T) {
			var calls, metadata atomic.Int64
			r, _ := setup(t, true, `["medium"]`, func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/generation" {
					metadata.Add(1)
					http.Error(w, "unavailable", http.StatusServiceUnavailable)
					return
				}
				calls.Add(1)
				id := "gen-unavailable"
				if missingID {
					id = ""
				}
				fmt.Fprintf(w, `{"id":%q,"choices":[{"finish_reason":"stop","message":{"content":"{\"answer\":4}"}}]}`, id)
			})
			snap := snapshot()
			q := snap.Questions[0]
			q.Number, q.Text = 2, "another question"
			snap.Questions = append(snap.Questions, q)
			result, err := r.Run(context.Background(), snap, Options{Concurrency: 1, BudgetUSD: 1, Selection: storage.Selection{Efforts: []string{"medium"}}})
			if !errors.Is(err, ErrBudget) || calls.Load() != 1 || result.NewCost != nil || result.UnknownCostAttempts != 1 {
				t.Fatalf("result=%+v calls=%d err=%v", result, calls.Load(), err)
			}
			if missingID && metadata.Load() != 0 || !missingID && (metadata.Load() < 2 || metadata.Load() > 30) {
				t.Fatalf("metadata retries were not bounded: %d", metadata.Load())
			}
		})
	}
}

func TestBudgetWaitsForAllConcurrentAccounting(t *testing.T) {
	for _, outcome := range []string{"continue", "exhausted", "unrecoverable"} {
		t.Run(outcome, func(t *testing.T) {
			budget := &spendingBudget{limit: 1}
			for range 2 {
				if !budget.record(openrouter.HTTPResult{GenerationID: "gen-pending", Error: "stream interrupted"}, "unknown") {
					t.Fatal("unknown accounting was not registered")
				}
			}
			budget.settle(0.2, nil)
			// Resolving one of two charges must not reopen dispatch.
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			if err := budget.check(ctx); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("dispatch did not wait for both charges: %v", err)
			}
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			results := make(chan error, 8)
			for range 8 {
				go func() { results <- budget.check(ctx) }()
			}
			switch outcome {
			case "continue":
				budget.settle(0.3, nil)
			case "exhausted":
				budget.settle(0.8, nil)
			case "unrecoverable":
				budget.settle(0, errors.New("missing generation ID"))
			}
			for range 8 {
				err := <-results
				if outcome == "continue" && err != nil || outcome != "continue" && !errors.Is(err, ErrBudget) {
					t.Fatalf("waiter did not receive accounting outcome: %v", err)
				}
			}
		})
	}
}

func TestCancellationInterruptsAutomaticAccounting(t *testing.T) {
	var calls, metadata atomic.Int64
	r, _ := setup(t, true, `["medium"]`, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/generation" {
			metadata.Add(1)
			http.Error(w, "not ready", http.StatusNotFound)
			return
		}
		calls.Add(1)
		fmt.Fprint(w, `{"id":"gen-cancel","choices":[{"finish_reason":"stop","message":{"content":"{\"answer\":4}"}}]}`)
	})
	waiting := make(chan struct{})
	r.Sleep = func(ctx context.Context, _ time.Duration) error {
		close(waiting)
		<-ctx.Done()
		return ctx.Err()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan error, 1)
	go func() {
		_, err := r.Run(ctx, snapshot(), Options{Concurrency: 1, BudgetUSD: 1, Selection: storage.Selection{Efforts: []string{"medium"}}})
		finished <- err
	}()
	select {
	case <-waiting:
	case <-time.After(2 * time.Second):
		t.Fatal("automatic accounting did not start")
	}
	cancel()
	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) || calls.Load() != 1 || metadata.Load() != 1 {
			t.Fatalf("cancel failed: err=%v calls=%d metadata=%d", err, calls.Load(), metadata.Load())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("accounting wait ignored cancellation")
	}
}
