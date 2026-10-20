package bench

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"konkour-bench/internal/backbone"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

func Reconcile(ctx context.Context, s *storage.Store, c *openrouter.Client) (int, error) {
	if _, err := s.RecoverStale(); err != nil {
		return 0, err
	}
	attempts, err := s.Attempts("", "")
	if err != nil {
		return 0, err
	}
	n := 0
	var failures []error
	for _, a := range attempts {
		if a.Finished == "" || a.Details.GenerationID == "" {
			continue
		}
		run, err := s.Run(a.RunID)
		if err != nil {
			return n, err
		}
		if backbone.IsClaude(run.BaseURL) || a.Details.Backend == "claude" {
			continue
		}
		u := a.Usage()
		if u.Cost != nil && u.Prompt != nil && u.Completion != nil && u.Total != nil {
			continue
		}
		if err := reconcileAttempt(ctx, s, c, a, ""); err != nil {
			failures = append(failures, err)
		} else {
			n++
		}
		if ctx.Err() != nil {
			break
		}
	}
	return n, errors.Join(failures...)
}

// ReconcileID recovers historical requests whose IDs were supplied from Activity.
// It verifies model, creation time, and ID before attaching any external data.
func ReconcileID(ctx context.Context, s *storage.Store, c *openrouter.Client, attempt int64, id string) error {
	if attempt <= 0 || !strings.HasPrefix(id, "gen-") {
		return fmt.Errorf("reconcile requires a positive --attempt and a gen- --generation-id")
	}
	if _, err := s.RecoverStale(); err != nil {
		return err
	}
	attempts, err := s.Attempts("", "")
	if err != nil {
		return err
	}
	for _, a := range attempts {
		if a.ID == attempt {
			if a.Finished == "" {
				return fmt.Errorf("attempt %d is still owned or unverified", attempt)
			}
			if a.Details.GenerationID != "" && a.Details.GenerationID != id {
				return fmt.Errorf("attempt already has a different generation ID")
			}
			return reconcileAttempt(ctx, s, c, a, id)
		}
	}
	return fmt.Errorf("attempt %d not found", attempt)
}

func reconcileAttempt(ctx context.Context, s *storage.Store, c *openrouter.Client, a storage.Attempt, manualID string) error {
	id := a.Details.GenerationID
	if manualID != "" {
		id = manualID
	}
	run, err := s.Run(a.RunID)
	if err != nil {
		return err
	}
	if backbone.IsClaude(run.BaseURL) || a.Details.Backend == "claude" {
		return fmt.Errorf("Claude CLI attempts cannot be reconciled through OpenRouter")
	}
	client := *c
	client.BaseURL = backbone.EndpointFor(run.BaseURL, "")
	// Metadata requests must not hang an offline reconciliation indefinitely.
	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	response := client.Generation(callCtx, id)
	if response.Error != "" || response.Status != 200 {
		return fmt.Errorf("reconcile attempt %d (%s): HTTP %d %s %s", a.ID, id, response.Status, response.Error, response.Body)
	}
	var envelope struct {
		Data struct {
			ID           json.RawMessage `json:"id"`
			GenerationID string          `json:"generation_id"`
			RequestID    string          `json:"request_id"`
			Model        string          `json:"model"`
			Created      string          `json:"created_at"`
			Cost         *float64        `json:"total_cost"`
			Charge       *float64        `json:"usage"`
			Prompt       *int64          `json:"native_tokens_prompt"`
			Completion   *int64          `json:"native_tokens_completion"`
			Reasoning    *int64          `json:"native_tokens_reasoning"`
			Cached       *int64          `json:"native_tokens_cached"`
			Provider     string          `json:"provider_name"`
			Cancelled    bool            `json:"cancelled"`
			Finish       string          `json:"finish_reason"`
			Responses    []struct {
				Status int `json:"status"`
			} `json:"provider_responses"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body, &envelope); err != nil {
		return err
	}
	d := envelope.Data
	gotID := d.GenerationID
	if gotID == "" {
		_ = json.Unmarshal(d.ID, &gotID)
	}
	if gotID != id {
		return fmt.Errorf("generation ID mismatch for attempt %d", a.ID)
	}
	if a.Details.RequestID != "" && d.RequestID != "" && a.Details.RequestID != d.RequestID {
		return fmt.Errorf("request ID mismatch for attempt %d", a.ID)
	}
	if manualID != "" {
		req, err := s.Request(a.Hash)
		if err != nil {
			return err
		}
		var body openrouter.Request
		if err := json.Unmarshal(req.Body, &body); err != nil {
			return err
		}
		model := strings.TrimSuffix(body.Model, ":exacto")
		if d.Model != model && !strings.HasPrefix(d.Model, model+"-") {
			return fmt.Errorf("generation model does not match attempt %d", a.ID)
		}
		created, err := time.Parse(time.RFC3339Nano, d.Created)
		if err != nil {
			return fmt.Errorf("generation creation time missing or invalid")
		}
		started, err := time.Parse(time.RFC3339Nano, a.Started)
		if err != nil {
			return err
		}
		if delta := created.Sub(started); delta < -time.Minute || delta > time.Minute {
			return fmt.Errorf("generation creation time does not match attempt %d", a.ID)
		}
	}
	terminal := d.Cancelled || d.Finish != ""
	for _, r := range d.Responses {
		terminal = terminal || r.Status >= 400
	}
	if !terminal {
		return fmt.Errorf("generation %s has no terminal outcome yet; accounting not finalized", id)
	}
	u := a.Usage()
	if d.Cost == nil {
		d.Cost = d.Charge
	}
	if d.Cost != nil && (math.IsNaN(*d.Cost) || math.IsInf(*d.Cost, 0) || *d.Cost < 0) {
		return fmt.Errorf("invalid generation cost")
	}
	for _, v := range []*int64{d.Prompt, d.Completion, d.Reasoning, d.Cached} {
		if v != nil && *v < 0 {
			return fmt.Errorf("invalid generation token count")
		}
	}
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
	if u.PromptDetails.Cached == nil {
		u.PromptDetails.Cached = d.Cached
	}
	if u.Total == nil && u.Prompt != nil && u.Completion != nil {
		v := *u.Prompt + *u.Completion
		u.Total = &v
	}
	a.Details.GenerationID = id
	if d.RequestID != "" {
		a.Details.RequestID = d.RequestID
	}
	a.Details.Provider = d.Provider
	return s.ReconcileAttempt(a, u, response.Body)
}
