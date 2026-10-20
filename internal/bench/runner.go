// Package bench schedules one independent generation per question/model/effort.
package bench

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"konkour-bench/internal/backbone"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

func Protocol() string {
	return dataset.Hash([]string{"konkour-bench-v1", openrouter.Prompt, "minimal:low;medium;max:xhigh:high", "provider-defaults", "strict-answer-v1"})
}
func ProtocolFor(endpoint string) string {
	if backbone.IsMixed(endpoint) {
		return dataset.Hash([]string{Protocol(), "mixed-cli-v1"})
	}
	if backbone.IsClaude(endpoint) {
		return dataset.Hash([]string{Protocol(), "claude-cli-v1"})
	}
	return Protocol()
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
	BudgetUSD    float64
	WarnAfter    time.Duration
	budget       *spendingBudget
}
type Runner struct {
	Store   *storage.Store
	Client  *openrouter.Client
	Backend backbone.Client
	Log     *slog.Logger
	// Sleep is injectable so retry tests do not wait for wall-clock time.
	Sleep func(context.Context, time.Duration) error
}
type Summary struct {
	SubscriptionAttempts                                        int     `json:"subscription_attempts,omitempty"`
	Backend                                                     string  `json:"backend,omitempty"`
	EstimatedCost                                               float64 `json:"estimated_api_cost_usd,omitempty"`
	RunID                                                       string  `json:"run_id"`
	Complete, Cached, Unsupported, Skipped, Failed, NewAttempts int
	NewCost                                                     *float64 `json:"new_cost_usd"`
	KnownCost                                                   float64  `json:"known_cost_usd"`
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
	if r.Backend == nil && r.Client != nil {
		r.Backend = r.Client
	}
	if r.Backend == nil {
		return Summary{}, fmt.Errorf("inference backend is required")
	}
	if backbone.IsClaude(r.Backend.Endpoint()) && opts.BudgetUSD != 0 {
		return Summary{}, fmt.Errorf("--budget-usd does not apply to subscription usage; use 0 for Claude CLI")
	}
	if r.Log == nil {
		r.Log = slog.Default()
	}
	if r.Sleep == nil {
		r.Sleep = Wait
	}
	if opts.BudgetUSD < 0 || math.IsNaN(opts.BudgetUSD) || math.IsInf(opts.BudgetUSD, 0) || opts.Timeout < 0 || opts.WarnAfter < 0 {
		return Summary{}, fmt.Errorf("invalid budget or duration")
	}
	opts.budget = &spendingBudget{limit: opts.BudgetUSD}
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
	if _, err := r.Store.RecoverStale(); err != nil {
		return Summary{}, err
	}
	lock, err := r.Store.AcquireRunner()
	if err != nil {
		return Summary{}, err
	}
	defer lock.Close()
	cat, err := r.Backend.Discover(ctx, sel.Models)
	if err != nil {
		return Summary{}, err
	}
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return Summary{}, err
	}
	run := storage.Run{ID: time.Now().UTC().Format("20060102T150405") + "-" + hex.EncodeToString(random[:]), Created: time.Now().UTC(), Protocol: ProtocolFor(r.Backend.Endpoint()), Snapshot: snapshot, Catalog: cat, Selection: sel, BaseURL: r.Backend.Endpoint(), Timeout: opts.Timeout, BudgetUSD: opts.BudgetUSD}
	if err := r.Store.AddRun(run); err != nil {
		return Summary{}, err
	}
	worker, err := r.Store.StartWorker(run.ID)
	if err != nil {
		return Summary{}, err
	}
	defer worker.Close()
	if _, err := r.Store.RecoverStale(); err != nil {
		return Summary{}, err
	}
	defer func() {
		if err := r.Store.Worker(run.ID, "stopped"); err != nil {
			r.Log.Error("record runner stop", "error", err)
		}
	}()
	r.Log.Info("benchmark started", "run", run.ID, "models", len(sel.Models), "questions", len(questions), "efforts", sel.Efforts)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sem := make(chan struct{}, opts.Concurrency)
	claudeSem := make(chan struct{}, 1)
	claudeCtx, cancelClaude := context.WithCancel(ctx)
	defer cancelClaude()
	apiCtx, cancelAPI := context.WithCancel(ctx)
	defer cancelAPI()
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	summary := Summary{RunID: run.ID, Backend: backbone.Name(run.BaseURL)}
	for _, modelID := range sel.Models {
		wg.Add(1)
		go func(modelID string) {
			defer wg.Done()
			endpoint := backbone.EndpointFor(run.BaseURL, modelID)
			isCLI := backbone.IsClaude(endpoint)
			workCtx := ctx
			workOpts := opts
			if backbone.IsMixed(run.BaseURL) {
				if isCLI {
					workCtx = claudeCtx
					workOpts.budget = nil
				} else {
					workCtx = apiCtx
				}
			}
			model, exists := cat.Models[strings.TrimSuffix(modelID, ":exacto")]
			for _, q := range questions {
				for _, tier := range sel.Efforts {
					if workCtx.Err() != nil {
						return
					}
					j := storage.Job{RunID: run.ID, Question: q.ID(), Model: modelID, Tier: tier, Backend: backbone.Name(endpoint)}
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
						limitCLI := backbone.IsMixed(run.BaseURL) && isCLI
						if limitCLI {
							select {
							case claudeSem <- struct{}{}:
							case <-workCtx.Done():
								return
							}
						}
						select {
						case sem <- struct{}{}:
						case <-workCtx.Done():
							if limitCLI {
								<-claudeSem
							}
							return
						}
						var callErr error
						j, callErr = r.answer(workCtx, q, j, candidates, workOpts)
						if backbone.IsMixed(run.BaseURL) && j.Status == "fatal" {
							if isCLI {
								cancelClaude()
							} else {
								cancelAPI()
							}
						}
						<-sem
						if limitCLI {
							<-claudeSem
						}
						if callErr != nil {
							if errors.Is(callErr, context.Canceled) && ctx.Err() == nil && workCtx.Err() != nil {
								return
							}
							mu.Lock()
							if firstErr == nil {
								firstErr = callErr
							}
							mu.Unlock()
							if !errors.Is(callErr, ErrBudget) {
								cancel()
							}
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
							firstErr = fmt.Errorf("inference backend stopped dispatch: %s", j.Note)
						}
						mu.Unlock()
						if backbone.IsMixed(run.BaseURL) {
							if isCLI {
								cancelClaude()
							} else {
								cancelAPI()
							}
						} else {
							cancel()
						}
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
		u := a.Usage()
		if u.EstimatedCost != nil {
			summary.EstimatedCost += *u.EstimatedCost
		}
		if a.Details.Backend == "claude" {
			summary.SubscriptionAttempts++
			if backbone.IsMixed(run.BaseURL) {
				continue
			}
		}
		if u.Cost != nil {
			cost += *u.Cost
		} else {
			summary.UnknownCostAttempts++
		}
	}
	summary.KnownCost = cost
	if summary.UnknownCostAttempts == 0 {
		summary.NewCost = &cost
	}
	if firstErr != nil {
		return summary, firstErr
	}
	if ctx.Err() != nil {
		return summary, ctx.Err()
	}
	if summary.UnknownCostAttempts > 0 {
		if err := opts.budget.check(ctx); err != nil {
			return summary, err
		}
	}
	if summary.Failed > 0 {
		return summary, fmt.Errorf("run has %d unresolved failures; completed generations are cached", summary.Failed)
	}
	return summary, nil
}

func (r *Runner) answer(ctx context.Context, q dataset.Question, j storage.Job, candidates []string, opts Options) (storage.Job, error) {
	endpoint := backbone.EndpointFor(r.Backend.Endpoint(), j.Model)
	for _, effort := range candidates {
		j.Effort = effort
		body, err := backbone.Build(r.Backend.Endpoint(), q, j.Model, effort, true)
		if err != nil {
			return j, err
		}
		j.Hash = RequestHash(endpoint, body)
		if cached, err := r.Store.Request(j.Hash); err == nil && cached.State == "complete" {
			j.Status = "complete"
			j.Cached = true
			return j, nil
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return j, err
		}
		// Preserve historical completed generations. An uncertain legacy request
		// must not be bypassed merely by switching transport to streaming.
		legacy, err := backbone.Build(r.Backend.Endpoint(), q, j.Model, effort, false)
		if err != nil {
			return j, err
		}
		legacyHash := RequestHash(endpoint, legacy)
		old, oldErr := r.Store.Request(legacyHash)
		if oldErr != nil && !errors.Is(oldErr, sql.ErrNoRows) {
			return j, oldErr
		}
		if oldErr == nil && legacyHash != j.Hash {
			switch old.State {
			case "complete":
				j.Hash = legacyHash
				j.Status = "complete"
				j.Cached = true
				return j, nil
			case "rejected":
				continue
			case "pending":
				j.Hash = legacyHash
				j.Status = "infra_error"
				j.Note = "legacy request still owned or unverified"
				return j, nil
			case "unknown":
				if !opts.RetryUnknown {
					j.Hash = legacyHash
					j.Status = "infra_error"
					j.Note = "legacy outcome unknown; explicit retry required"
					return j, nil
				}
			}
		}
		for attempt := 0; attempt < 3; attempt++ {
			if cached, err := r.Store.Request(j.Hash); err == nil && cached.State == "complete" {
				j.Status = "complete"
				j.Cached = true
				return j, nil
			} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return j, err
			}
			if err := opts.budget.check(ctx); err != nil {
				return j, err
			}
			if err := ctx.Err(); err != nil {
				return j, err
			}
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
			if err := r.Store.StartAttempt(req, j); err != nil {
				return j, err
			}
			j.AttemptID = req.AttemptID
			j.Status = "pending"
			if err := r.Store.AddJob(j); err != nil {
				return j, err
			}
			callCtx := ctx
			cancel := func() {}
			if opts.Timeout > 0 {
				callCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
			}
			var warning *time.Timer
			if opts.WarnAfter > 0 {
				warning = time.AfterFunc(opts.WarnAfter, func() {
					r.Log.Warn("request taking longer than warning threshold", "attempt", req.AttemptID, "question", q.ID(), "model", j.Model, "tier", j.Tier, "threshold", opts.WarnAfter.String())
				})
			}
			response := r.Backend.Complete(callCtx, body, func(p openrouter.Progress) error { return r.Store.Progress(req, p) })
			deadline := errors.Is(callCtx.Err(), context.DeadlineExceeded)
			if warning != nil {
				warning.Stop()
			}
			cancel()
			state, note := "retryable", ""
			var result storage.Result
			if response.GenerationID != "" {
				result.Generation.ID = response.GenerationID
			}
			switch {
			case response.Error != "":
				state = "unknown"
				note = response.Error
				if deadline {
					note = "deadline_exceeded: provider outcome unknown"
				}
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
			if state != "complete" {
				r.Log.Warn("inference attempt failed", "question", q.ID(), "model", j.Model, "tier", j.Tier, "effort", effort, "attempt", attempt+1, "status", response.Status, "state", state, "error", note)
			}
			if opts.budget.record(response, state) {
				r.Log.Info("waiting for generation accounting; new dispatch paused", "attempt", req.AttemptID, "generation", response.GenerationID)
				cost, err := r.reconcileCost(ctx, req, j.RunID)
				opts.budget.settle(cost, err)
				if err != nil {
					r.Log.Error("automatic accounting recovery failed", "attempt", req.AttemptID, "error", err)
				} else {
					r.Log.Info("generation accounting recovered", "attempt", req.AttemptID, "cost_usd", cost)
				}
			}
			if state == "complete" {
				j.Status = "complete"
				return j, nil
			}
			if state == "rejected" {
				break
			}
			j.Status = "infra_error"
			j.Note = note
			if response.Fatal || response.Status == 401 || response.Status == 402 || response.Status == 403 {
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
