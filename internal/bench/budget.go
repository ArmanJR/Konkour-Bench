package bench

import (
	"encoding/json"
	"errors"
	"fmt"
	"konkour-bench/internal/openrouter"
	"sync"
)

var ErrBudget = errors.New("dispatch spending budget stopped")

// This is a dispatch threshold, not a hard billing cap. Already dispatched
// requests can overshoot it. Unknown accounting stops further dispatch as well.
type spendingBudget struct {
	mu           sync.Mutex
	limit, spent float64
	unknown      bool
}

func (b *spendingBudget) check() error {
	if b == nil || b.limit == 0 {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.unknown {
		return fmt.Errorf("%w: an attempt has unknown cost; reconcile before resuming", ErrBudget)
	}
	if b.spent >= b.limit {
		return fmt.Errorf("%w: recorded $%.6f reached $%.6f", ErrBudget, b.spent, b.limit)
	}
	return nil
}
func (b *spendingBudget) record(r openrouter.HTTPResult, state string) {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	var g openrouter.Generation
	if state == "complete" && json.Unmarshal(r.Body, &g) == nil && g.Usage.Cost != nil {
		b.spent += *g.Usage.Cost
		return
	}
	// Explicit pre-generation validation/rate/account rejections have no output.
	if r.GenerationID == "" && r.Error == "" && (r.Status == 400 || r.Status == 401 || r.Status == 402 || r.Status == 403 || r.Status == 422 || r.Status == 429) {
		return
	}
	b.unknown = true
}
