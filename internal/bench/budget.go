package bench

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"

	"konkour-bench/internal/openrouter"
)

var ErrBudget = errors.New("dispatch spending budget stopped")

// This is a dispatch threshold, not a hard billing cap. Already dispatched
// requests can overshoot it. Dispatch waits for unresolved accounting, while
// the goroutine that recorded each attempt reconciles it independently.
type spendingBudget struct {
	mu           sync.Mutex
	limit, spent float64
	pending      int
	changed      chan struct{}
	failure      error
}

func (b *spendingBudget) check(ctx context.Context) error {
	if b == nil || b.limit == 0 {
		return nil
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		b.mu.Lock()
		err := b.failure
		if err == nil && b.spent >= b.limit {
			err = fmt.Errorf("%w: recorded $%.6f reached $%.6f", ErrBudget, b.spent, b.limit)
		}
		if err != nil || b.pending == 0 {
			b.mu.Unlock()
			return err
		}
		if b.changed == nil {
			b.changed = make(chan struct{})
		}
		changed := b.changed
		b.mu.Unlock()
		select {
		case <-changed:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// record returns true when the caller must settle this attempt's accounting.
func (b *spendingBudget) record(r openrouter.HTTPResult, state string) bool {
	if b == nil || b.limit == 0 {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	var g openrouter.Generation
	if state == "complete" && json.Unmarshal(r.Body, &g) == nil && validCost(g.Usage.Cost) {
		b.spent += *g.Usage.Cost
		b.notify()
		return false
	}
	// Explicit pre-generation validation/rate/account rejections have no output.
	if r.GenerationID == "" && r.Error == "" && (r.Status == 400 || r.Status == 401 || r.Status == 402 || r.Status == 403 || r.Status == 422 || r.Status == 429) {
		return false
	}
	b.pending++
	return true
}

func validCost(cost *float64) bool {
	return cost != nil && !math.IsNaN(*cost) && !math.IsInf(*cost, 0) && *cost >= 0
}

func (b *spendingBudget) settle(cost float64, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pending--
	if err != nil {
		if b.failure == nil {
			b.failure = fmt.Errorf("%w: automatic accounting recovery failed: %v", ErrBudget, err)
		}
	} else {
		b.spent += cost
	}
	b.notify()
}

// notify is called with mu held; every waiter rechecks all outstanding costs.
func (b *spendingBudget) notify() {
	if b.changed != nil {
		close(b.changed)
		b.changed = nil
	}
}
