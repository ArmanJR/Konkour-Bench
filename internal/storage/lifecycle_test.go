package storage

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"konkour-bench/internal/openrouter"
)

func TestDeadWorkerRecoveryAndExplicitRetry(t *testing.T) {
	s := testStore(t)
	lock, err := s.AcquireRunner()
	if err != nil {
		t.Fatal(err)
	}
	worker, err := s.StartWorker("run")
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := s.Claim("run", "hash", []byte(`{"model":"test/model"}`), false)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.StartAttempt(r, Job{Question: "exam:1", Tier: "max"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Progress(r, openrouter.Progress{GenerationID: "gen-partial", RequestID: "req-partial", Raw: `{"choices":[]}`, HTTPStatus: 200}); err != nil {
		t.Fatal(err)
	}
	active, err := s.Activity("")
	if err != nil || len(active) != 1 || active[0].LocalState != "running" {
		t.Fatalf("active %+v %v", active, err)
	}
	// Simulate process death: the kernel lock is released, even though PID exists.
	lock.Close()
	worker.Close()
	active, err = s.Activity("")
	if err != nil || len(active) != 1 || active[0].State != "unknown" || active[0].LocalState != "outcome_unknown" || active[0].Details.GenerationID != "gen-partial" {
		t.Fatalf("stale %+v %v", active, err)
	}
	_, claimed, err := s.Claim("run", "hash", r.Body, false)
	if err != nil || claimed {
		t.Fatalf("uncertain request replayed: %v %v", claimed, err)
	}
	retry, claimed, err := s.Claim("run", "hash", r.Body, true)
	if err != nil || !claimed || retry.AttemptID == r.AttemptID || retry.Result.Generation.ID != "" {
		t.Fatalf("retry %+v %v %v", retry, claimed, err)
	}
	attempts, _ := s.Attempts("hash", "")
	if attempts[0].Error == "" || attempts[0].Details.GenerationID != "gen-partial" {
		t.Fatal("lost interrupted audit", attempts)
	}
	events, err := s.Events(r.AttemptID)
	if err != nil || len(events) != 1 {
		t.Fatalf("events %v %v", events, err)
	}
}

func TestLegacyPendingIsNotClaimedActive(t *testing.T) {
	s := testStore(t)
	r, _, _ := s.Claim("run", "hash", []byte(`{}`), false)
	activity, err := s.Activity("")
	if err != nil || activity[0].LocalState != "unverified" {
		t.Fatalf("legacy %+v %v", activity, err)
	}
	if _, err := s.db.Exec("UPDATE requests SET owner=0 WHERE hash=?", r.Hash); err != nil {
		t.Fatal(err)
	}
	req, claimed, err := s.Claim("run", r.Hash, r.Body, false)
	if err != nil || claimed || req.State != "unknown" {
		t.Fatalf("claim %+v %v %v", req, claimed, err)
	}
	req, err = s.Request(r.Hash)
	if err != nil || req.State != "unknown" {
		t.Fatal("dead legacy state not persisted", req, err)
	}
	attempts, err := s.Attempts(r.Hash, "")
	if err != nil || attempts[0].Finished == "" {
		t.Fatal("orphaned attempt cannot be reconciled", attempts, err)
	}
}

func TestRunnerLockAndMigrationReopen(t *testing.T) {
	s := testStore(t)
	other, err := Open(s.path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	lock, err := s.AcquireRunner()
	if err != nil {
		t.Fatal(err)
	}
	if second, err := other.AcquireRunner(); err == nil {
		second.Close()
		t.Fatal("second runner admitted")
	}
	lock.Close()
	next, err := other.AcquireRunner()
	if err != nil {
		t.Fatal(err)
	}
	next.Close()
}

func TestReconciliationBelongsToAttemptNotRetry(t *testing.T) {
	s := testStore(t)
	old, _, _ := s.Claim("run", "hash", []byte(`{}`), false)
	if err := s.Finish(old, "unknown", openrouter.HTTPResult{Error: "disconnected"}, Result{}, "unknown"); err != nil {
		t.Fatal(err)
	}
	retry, _, _ := s.Claim("run", "hash", []byte(`{}`), true)
	if err := s.Finish(retry, "complete", openrouter.HTTPResult{Status: 200, Body: []byte(`{"id":"gen-new","usage":{"cost":0.2}}`)}, Result{Generation: openrouter.Generation{ID: "gen-new"}, Answer: 4}, ""); err != nil {
		t.Fatal(err)
	}
	attempts, _ := s.Attempts("hash", "")
	a := attempts[0]
	a.Details.GenerationID = "gen-old"
	cost := 0.1
	u := openrouter.Usage{Cost: &cost}
	if err := s.ReconcileAttempt(a, u, []byte(`{"data":{"id":"gen-old"}}`)); err != nil {
		t.Fatal(err)
	}
	req, _ := s.Request("hash")
	if req.Result.Answer != 4 || req.Result.Generation.ID != "gen-new" || req.Result.Generation.Usage.Cost != nil {
		t.Fatal("old accounting changed latest answer", req)
	}
	attempts, _ = s.Attempts("hash", "")
	if *attempts[0].Usage().Cost != 0.1 || *attempts[1].Usage().Cost != 0.2 {
		t.Fatal("incorrect attempt ledger", attempts)
	}
}

func TestMigrateRealV1Schema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "v1.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	// Reconstruct the original schema instead of editing generated data fixtures.
	_, err = s.db.Exec(`DROP TABLE stream_events; DROP TABLE workers; ALTER TABLE attempts DROP COLUMN details; ALTER TABLE reconciliations DROP COLUMN attempt_id; PRAGMA user_version=1;`)
	if err != nil {
		t.Fatal(err)
	}
	s.AddRun(Run{ID: "old", Protocol: "old"})
	_, err = s.db.Exec(`INSERT INTO requests(hash,body,state,result,owner) VALUES('old','{}','complete','{"answer":3}',1)`)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	migrated, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer migrated.Close()
	req, err := migrated.Request("old")
	if err != nil || req.Result.Answer != 3 {
		t.Fatalf("migration lost data %+v %v", req, err)
	}
	var v int
	if err := migrated.db.QueryRow("PRAGMA user_version").Scan(&v); err != nil || v != 2 {
		t.Fatal(v, err)
	}
}

func TestProgressRejectsDifferentGeneration(t *testing.T) {
	s := testStore(t)
	r, _, _ := s.Claim("run", "hash", []byte(`{}`), false)
	if err := s.Progress(r, openrouter.Progress{GenerationID: "gen-one"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Progress(r, openrouter.Progress{GenerationID: "gen-two"}); err == nil {
		t.Fatal("changed ID accepted")
	}
	attempts, _ := s.Attempts("hash", "")
	b, _ := json.Marshal(attempts)
	if len(b) == 0 || attempts[0].Details.GenerationID != "gen-one" {
		t.Fatal("audit overwritten")
	}
	if _, err := os.Stat(s.path); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerCrashProcess(t *testing.T) {
	if path := os.Getenv("KONKOUR_TEST_CRASH_DB"); path != "" {
		s, err := Open(path)
		if err != nil {
			panic(err)
		}
		if err := s.AddRun(Run{ID: "crash", Protocol: "test"}); err != nil {
			panic(err)
		}
		lock, err := s.AcquireRunner()
		if err != nil {
			panic(err)
		}
		worker, err := s.StartWorker("crash")
		if err != nil {
			panic(err)
		}
		if _, _, err := s.Claim("crash", "hash", []byte(`{"model":"test/model"}`), false); err != nil {
			panic(err)
		}
		// Exit without cleanup or a final response. Retain the file handle until exit.
		runtime.KeepAlive(lock)
		runtime.KeepAlive(worker)
		os.Exit(0)
	}
	path := filepath.Join(t.TempDir(), "crash.db")
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary, "-test.run=^TestWorkerCrashProcess$")
	cmd.Env = append(os.Environ(), "KONKOUR_TEST_CRASH_DB="+path)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("helper: %v %s", err, output)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	activity, err := s.Activity("")
	if err != nil || len(activity) != 1 || activity[0].LocalState != "outcome_unknown" {
		t.Fatalf("crashed worker still active: %+v %v", activity, err)
	}
}

func TestUnrelatedRunnerLockDoesNotReviveDeadWorker(t *testing.T) {
	s := testStore(t)
	worker, err := s.StartWorker("run")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Claim("run", "hash", []byte(`{}`), false); err != nil {
		t.Fatal(err)
	}
	worker.Close()
	// A newly starting process may hold the database dispatch lock while doing
	// discovery. That is not proof that the previous run is still alive.
	newRunner, err := s.AcquireRunner()
	if err != nil {
		t.Fatal(err)
	}
	defer newRunner.Close()
	activity, err := s.Activity("")
	if err != nil || len(activity) != 1 || activity[0].LocalState != "outcome_unknown" {
		t.Fatalf("old worker revived: %+v %v", activity, err)
	}
}
