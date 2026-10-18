// Package bench schedules one independent generation per question/model/effort.
package bench

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

func Protocol() string {
	return dataset.Hash([]string{"konkour-bench-v1", openrouter.Prompt, "minimal:low;medium;max:xhigh:high", "provider-defaults", "strict-answer-v1"})
}
func RequestHash(base string, body []byte) string {
	return dataset.Hash(struct {
		URL  string
		Body json.RawMessage
	}{strings.TrimRight(base, "/") + "/chat/completions", body})
}

type Options struct {
	Selection    storage.Selection
	Concurrency  int
	Timeout      time.Duration
	RetryUnknown bool
}
type Runner struct {
	Store  *storage.Store
	Client *openrouter.Client
	Log    *slog.Logger
	// Sleep is injectable so retry tests do not wait for wall-clock time.
	Sleep func(context.Context, time.Duration) error
}
type Summary struct {
	RunID                                                       string `json:"run_id"`
	Complete, Cached, Unsupported, Skipped, Failed, NewAttempts int
	NewCost                                                     *float64 `json:"new_cost_usd"`
	UnknownCostAttempts                                         int      `json:"unknown_cost_attempts"`
}

func Wait(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (r *Runner) Run(ctx context.Context, snapshot dataset.Snapshot, opts Options) (Summary, error) {
	if r.Log == nil {
		r.Log = slog.Default()
	}
	if r.Sleep == nil {
		r.Sleep = Wait
	}
	if opts.Concurrency < 1 {
		return Summary{}, fmt.Errorf("concurrency must be positive")
	}
	sel := opts.Selection
	if len(sel.Models) == 0 {
		sel.Models = snapshot.Config.Models
	}
	if len(sel.Efforts) == 0 {
		sel.Efforts = openrouter.Tiers
	}
	for name, values := range map[string][]string{"model": sel.Models, "effort": sel.Efforts} {
		seen := map[string]bool{}
		for _, v := range values {
			if seen[v] {
				return Summary{}, fmt.Errorf("duplicate %s selector %q", name, v)
			}
			seen[v] = true
		}
	}
	for _, tier := range sel.Efforts {
		if !slices.Contains(openrouter.Tiers, tier) {
			return Summary{}, fmt.Errorf("unknown effort %q", tier)
		}
	}
	for _, id := range sel.Models {
		if !slices.Contains(snapshot.Config.Models, id) {
			return Summary{}, fmt.Errorf("model %q is not in configuration", id)
		}
	}
	var questions []dataset.Question
	for _, q := range snapshot.Questions {
		if len(sel.Questions) == 0 || slices.Contains(sel.Questions, q.ID()) {
			questions = append(questions, q)
		}
	}
	if len(questions) == 0 {
		return Summary{}, fmt.Errorf("no questions selected")
	}
	if len(sel.Questions) > 0 && len(questions) != len(sel.Questions) {
		return Summary{}, fmt.Errorf("unknown or duplicate question selector")
	}
	sel.Questions = nil
	for _, q := range questions {
		sel.Questions = append(sel.Questions, q.ID())
	}
	cat, err := r.Client.Discover(ctx, sel.Models)
	if err != nil {
		return Summary{}, err
	}
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return Summary{}, err
	}
	run := storage.Run{ID: time.Now().UTC().Format("20060102T150405") + "-" + hex.EncodeToString(random[:]), Created: time.Now().UTC(), Protocol: Protocol(), Snapshot: snapshot, Catalog: cat, Selection: sel, BaseURL: r.Client.BaseURL}
	if err := r.Store.AddRun(run); err != nil {
		return Summary{}, err
	}
	r.Log.Info("benchmark started", "run", run.ID, "models", len(sel.Models), "questions", len(questions), "efforts", sel.Efforts)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sem := make(chan struct{}, opts.Concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	summary := Summary{RunID: run.ID}
	for _, modelID := range sel.Models {
		wg.Add(1)
		go func(modelID string) {
			defer wg.Done()
			model, exists := cat.Models[strings.TrimSuffix(modelID, ":exacto")]
			for _, q := range questions {
				for _, tier := range sel.Efforts {
					if ctx.Err() != nil {
						return
					}
					j := storage.Job{RunID: run.ID, Question: q.ID(), Model: modelID, Tier: tier}
					candidates, evidence := model.Candidates(tier)
					switch {
					case !exists:
						j.Status = "unsupported"
						j.Note = "model absent from catalog"
					case q.HasImage() && !model.Images():
						j.Status = "skipped"
						j.Note = "model does not accept images"
					case len(candidates) == 0:
						j.Status = "unsupported"
						j.Note = evidence
					default:
						select {
						case sem <- struct{}{}:
						case <-ctx.Done():
							return
						}
						var callErr error
						j, callErr = r.answer(ctx, q, j, candidates, opts)
						<-sem
						if callErr != nil {
							mu.Lock()
							if firstErr == nil {
								firstErr = callErr
							}
							mu.Unlock()
							cancel()
							return
						}
						if j.Status == "complete" {
							j.Note = evidence
						}
					}
					if err := r.Store.AddJob(j); err != nil {
						mu.Lock()
						if firstErr == nil {
							firstErr = err
						}
						mu.Unlock()
						cancel()
						return
					}
					mu.Lock()
					switch j.Status {
					case "complete":
						summary.Complete++
						if j.Cached {
							summary.Cached++
						}
					case "unsupported":
						summary.Unsupported++
					case "skipped":
						summary.Skipped++
					default:
						summary.Failed++
					}
					mu.Unlock()
					r.Log.Info("question evaluated", "run", run.ID, "question", q.ID(), "model", modelID, "tier", tier, "sent_effort", j.Effort, "status", j.Status, "cached", j.Cached, "detail", j.Note)
					if j.Status == "fatal" {
						mu.Lock()
						if firstErr == nil {
							firstErr = fmt.Errorf("OpenRouter authentication or account error: %s", j.Note)
						}
						mu.Unlock()
						cancel()
						return
					}
				}
			}
		}(modelID)
	}
	wg.Wait()
	attempts, err := r.Store.Attempts("", run.ID)
	if err != nil {
		return summary, err
	}
	summary.NewAttempts = len(attempts)
	var cost float64
	for _, a := range attempts {
		var g openrouter.Generation
		if json.Unmarshal([]byte(a.Response), &g) == nil && g.Usage.Cost != nil {
			cost += *g.Usage.Cost
		} else {
			summary.UnknownCostAttempts++
		}
	}
	if summary.UnknownCostAttempts == 0 {
		summary.NewCost = &cost
	}
	if firstErr != nil {
		return summary, firstErr
	}
	if ctx.Err() != nil {
		return summary, ctx.Err()
	}
	if summary.Failed > 0 {
		return summary, fmt.Errorf("run has %d unresolved failures; completed generations are cached", summary.Failed)
	}
	return summary, nil
}

func (r *Runner) answer(ctx context.Context, q dataset.Question, j storage.Job, candidates []string, opts Options) (storage.Job, error) {
	for _, effort := range candidates {
		j.Effort = effort
		body, err := openrouter.Build(q, j.Model, effort)
		if err != nil {
			return j, err
		}
		j.Hash = RequestHash(r.Client.BaseURL, body)
		for attempt := 0; attempt < 3; attempt++ {
			req, claimed, err := r.Store.Claim(j.RunID, j.Hash, body, opts.RetryUnknown && attempt == 0)
			if err != nil {
				return j, err
			}
			if !claimed {
				if req.State == "complete" {
					j.Status = "complete"
					j.Cached = true
					return j, nil
				}
				if req.State == "rejected" {
					break
				}
				j.Status = "infra_error"
				j.Note = req.State + ": " + req.Note
				return j, nil
			}
			callCtx := ctx
			cancel := func() {}
			if opts.Timeout > 0 {
				callCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
			}
			response := r.Client.Do(callCtx, http.MethodPost, "/chat/completions", body)
			cancel()
			state, note := "retryable", ""
			var result storage.Result
			switch {
			case response.Error != "":
				state = "unknown"
				note = response.Error
			case openrouter.EffortRejected(response):
				state = "rejected"
				note = string(response.Body)
			case response.Status == 200:
				g, answer, invalid, parseErr := openrouter.Parse(response.Body)
				result = storage.Result{Generation: g, Answer: answer, Invalid: invalid}
				if parseErr != nil {
					state = "unknown"
					note = parseErr.Error()
				} else {
					state = "complete"
					note = invalid
				}
			default:
				note = fmt.Sprintf("HTTP %d: %s", response.Status, response.Body)
			}
			if err := r.Store.Finish(req, state, response, result, note); err != nil {
				return j, err
			}
			if state == "complete" {
				j.Status = "complete"
				return j, nil
			}
			r.Log.Warn("inference attempt failed", "question", q.ID(), "model", j.Model, "tier", j.Tier, "effort", effort, "attempt", attempt+1, "status", response.Status, "state", state, "error", note)
			if state == "rejected" {
				break
			}
			j.Status = "infra_error"
			j.Note = note
			if response.Status == 401 || response.Status == 402 || response.Status == 403 {
				j.Status = "fatal"
				return j, nil
			}
			if state == "unknown" || (response.Status != 429 && response.Status < 500) || attempt == 2 {
				return j, nil
			}
			delay := time.Second * time.Duration(1<<attempt)
			if response.RetryAfter > delay {
				delay = response.RetryAfter
			}
			if err := r.Sleep(ctx, delay); err != nil {
				return j, err
			}
		}
	}
	j.Status = "unsupported"
	j.Note = "all allowed effort values explicitly rejected"
	return j, nil
}

func Reconcile(ctx context.Context, s *storage.Store, c *openrouter.Client) (int, error) {
	requests, err := s.MissingUsage()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range requests {
		response := c.Generation(ctx, r.Result.Generation.ID)
		if response.Error != "" || response.Status != 200 {
			return n, fmt.Errorf("reconcile %s: HTTP %d %s %s", r.Result.Generation.ID, response.Status, response.Error, response.Body)
		}
		var envelope struct {
			Data struct {
				Cost       *float64 `json:"total_cost"`
				Prompt     *int64   `json:"native_tokens_prompt"`
				Completion *int64   `json:"native_tokens_completion"`
				Reasoning  *int64   `json:"native_tokens_reasoning"`
				Provider   string   `json:"provider_name"`
			} `json:"data"`
		}
		if err := json.Unmarshal(response.Body, &envelope); err != nil {
			return n, err
		}
		u := &r.Result.Generation.Usage
		d := envelope.Data
		if u.Cost == nil {
			u.Cost = d.Cost
		}
		if u.Prompt == nil {
			u.Prompt = d.Prompt
		}
		if u.Completion == nil {
			u.Completion = d.Completion
		}
		if u.CompletionDetails.Reasoning == nil {
			u.CompletionDetails.Reasoning = d.Reasoning
		}
		if u.Total == nil && u.Prompt != nil && u.Completion != nil {
			v := *u.Prompt + *u.Completion
			u.Total = &v
		}
		if r.Result.Generation.Provider == "" {
			r.Result.Generation.Provider = d.Provider
		}
		if err := s.Reconcile(r, response.Body); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
