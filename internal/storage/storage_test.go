package storage

import (
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"konkour-bench/internal/openrouter"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.AddRun(Run{ID: "run", Protocol: "test"}); err != nil {
		t.Fatal(err)
	}
	return s
}
func TestAtomicClaimsAndReuse(t *testing.T) {
	s := testStore(t)
	var count atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, claimed, err := s.Claim("run", "hash", []byte(`{"model":"test"}`), false)
			if err != nil {
				t.Error(err)
			}
			if claimed {
				count.Add(1)
			}
		}()
	}
	wg.Wait()
	if count.Load() != 1 {
		t.Fatalf("%d calls claimed", count.Load())
	}
	r, err := s.Request("hash")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Finish(r, "complete", openrouter.HTTPResult{Status: 200, Body: []byte(`{"choices":[]}`)}, Result{Answer: 3}, ""); err != nil {
		t.Fatal(err)
	}
	r, claimed, err := s.Claim("run", "hash", r.Body, false)
	if err != nil || claimed || r.Result.Answer != 3 {
		t.Fatalf("cache result=%+v claimed=%v err=%v", r, claimed, err)
	}
	attempts, err := s.Attempts("hash", "")
	if err != nil || len(attempts) != 1 || attempts[0].Response != `{"choices":[]}` {
		t.Fatalf("audit records: %+v %v", attempts, err)
	}
}
func TestUnknownRequiresExplicitRetry(t *testing.T) {
	s := testStore(t)
	r, _, err := s.Claim("run", "hash", []byte(`{}`), false)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Finish(r, "unknown", openrouter.HTTPResult{Error: "connection lost"}, Result{}, "uncertain"); err != nil {
		t.Fatal(err)
	}
	_, claimed, err := s.Claim("run", "hash", r.Body, false)
	if err != nil || claimed {
		t.Fatalf("unknown replayed: %v %v", claimed, err)
	}
	_, claimed, err = s.Claim("run", "hash", r.Body, true)
	if err != nil || !claimed {
		t.Fatalf("explicit retry failed: %v %v", claimed, err)
	}
}
