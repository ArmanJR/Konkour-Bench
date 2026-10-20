package bench

import (
	"context"
	"fmt"
	"time"

	"konkour-bench/internal/storage"
)

// reconcileCost retries metadata reads, never inference. Accounting can lag a
// cancelled stream. Other requests may finish while dispatch waits for this
// charge; once all charges are known, waiting workers continue in the same run.
func (r *Runner) reconcileCost(ctx context.Context, req storage.Request, runID string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	load := func() (storage.Attempt, error) {
		attempts, err := r.Store.Attempts(req.Hash, runID)
		if err != nil {
			return storage.Attempt{}, err
		}
		for _, a := range attempts {
			if a.ID == req.AttemptID {
				return a, nil
			}
		}
		return storage.Attempt{}, fmt.Errorf("attempt %d not found", req.AttemptID)
	}
	var lastErr error
	for poll := 0; poll < 30; poll++ {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		a, err := load()
		if err != nil {
			return 0, err
		}
		if cost := a.Usage().Cost; validCost(cost) {
			return *cost, nil
		}
		if a.Details.GenerationID == "" {
			return 0, fmt.Errorf("attempt %d has unknown cost and no generation ID", a.ID)
		}
		lastErr = reconcileAttempt(ctx, r.Store, r.Client, a, "")
		if lastErr == nil {
			a, err = load()
			if err != nil {
				return 0, err
			}
			if cost := a.Usage().Cost; validCost(cost) {
				return *cost, nil
			}
			lastErr = fmt.Errorf("generation %s has no valid final cost yet", a.Details.GenerationID)
		}
		r.Log.Warn("generation accounting unavailable", "attempt", a.ID, "poll", poll+1, "error", lastErr)
		if poll < 29 {
			if err := r.Sleep(ctx, 10*time.Second); err != nil {
				return 0, err
			}
		}
	}
	return 0, fmt.Errorf("attempt %d accounting remains unavailable: %w", req.AttemptID, lastErr)
}
