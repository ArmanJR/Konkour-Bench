package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"konkour-bench/internal/openrouter"
)

func (s *Store) migrate() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Serialize migrations before reading the version, including concurrent opens.
	if _, err = tx.Exec("UPDATE runs SET protocol=protocol WHERE 0"); err != nil {
		return err
	}
	var version int
	if err = tx.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version > 2 {
		return fmt.Errorf("database schema %d is newer than this binary", version)
	}
	if version < 2 {
		if _, err = tx.Exec(`ALTER TABLE attempts ADD COLUMN details BLOB NOT NULL DEFAULT '{}';
ALTER TABLE reconciliations ADD COLUMN attempt_id INTEGER;
CREATE TABLE stream_events (id INTEGER PRIMARY KEY AUTOINCREMENT, attempt_id INTEGER NOT NULL REFERENCES attempts(id), received TEXT NOT NULL, data BLOB NOT NULL);
CREATE TABLE workers (run_id TEXT PRIMARY KEY REFERENCES runs(id), state TEXT NOT NULL, updated TEXT NOT NULL);
PRAGMA user_version=2;`); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// The kernel releases this lock on process death, unlike a stored PID, which can
// be stale or reused. Only one runner may dispatch work into a database at once.
func (s *Store) AcquireRunner() (*os.File, error) {
	return acquireFile(s.path + ".runner.lock")
}
func acquireFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("benchmark runner lock: %w", err)
	}
	return f, nil
}
func (s *Store) workerPath(runID string) (string, error) {
	if runID == "" || filepath.Base(runID) != runID || runID == "." || runID == ".." {
		return "", fmt.Errorf("invalid worker run ID")
	}
	dir := s.path + ".workers"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, runID+".lock"), nil
}
func (s *Store) workerLocked(runID string) (bool, error) {
	path, err := s.workerPath(runID)
	if err != nil {
		return false, err
	}
	f, err := acquireFile(path)
	if err == nil {
		return false, f.Close()
	}
	if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
		return true, nil
	}
	return false, err
}

func (s *Store) StartWorker(runID string) (*os.File, error) {
	path, err := s.workerPath(runID)
	if err != nil {
		return nil, err
	}
	f, err := acquireFile(path)
	if err != nil {
		return nil, err
	}
	if err := s.Worker(runID, "running"); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
func (s *Store) Worker(runID, state string) error {
	if state == "running" {
		if _, err := s.db.Exec("UPDATE workers SET state='interrupted' WHERE state='running' AND run_id<>?", runID); err != nil {
			return err
		}
	}
	_, err := s.db.Exec("INSERT INTO workers(run_id,state,updated) VALUES(?,?,?) ON CONFLICT(run_id) DO UPDATE SET state=excluded.state,updated=excluded.updated", runID, state, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

type Activity struct {
	Attempt
	Model                string  `json:"model"`
	Effort               string  `json:"sent_effort"`
	State                string  `json:"state"`
	LocalState           string  `json:"local_state"`
	SinceDispatchSeconds float64 `json:"seconds_since_dispatch"`
}

// RecoverStale never infers provider activity from elapsed time. Legacy claims
// with a live/inaccessible PID are unverified; new claims use the kernel lock.
func (s *Store) RecoverStale() (int, error) {
	probe, err := s.AcquireRunner()
	if err == nil {
		defer probe.Close()
	} else if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
		// Another runner owns dispatch. Its individual run lock determines liveness.
	} else {
		return 0, err
	}
	rows, err := s.db.Query(`SELECT r.hash,r.attempt_id,r.owner,coalesce(w.state,''),a.run_id FROM requests r JOIN attempts a ON a.id=r.attempt_id LEFT JOIN workers w ON w.run_id=a.run_id WHERE r.state='pending'`)
	if err != nil {
		return 0, err
	}
	type stale struct {
		hash    string
		attempt int64
	}
	var dead []stale
	for rows.Next() {
		var hash, worker, runID string
		var attempt int64
		var owner int
		if err := rows.Scan(&hash, &attempt, &owner, &worker, &runID); err != nil {
			rows.Close()
			return 0, err
		}
		locked := false
		if worker != "" {
			var err error
			locked, err = s.workerLocked(runID)
			if err != nil {
				rows.Close()
				return 0, err
			}
		}
		if (worker != "" && (!locked || worker != "running")) || (worker == "" && !processAlive(owner)) {
			dead = append(dead, stale{hash, attempt})
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, d := range dead {
		tx, err := s.db.Begin()
		if err != nil {
			return n, err
		}
		note := "runner exited without a final response; provider outcome unknown"
		res, err := tx.Exec("UPDATE requests SET state='unknown',note=? WHERE hash=? AND attempt_id=? AND state='pending'", note, d.hash, d.attempt)
		if err == nil {
			var count int64
			count, err = res.RowsAffected()
			if err == nil && count == 1 {
				_, err = tx.Exec("UPDATE attempts SET finished=?,error=? WHERE id=? AND finished=''", time.Now().UTC().Format(time.RFC3339Nano), note, d.attempt)
				n++
			}
		}
		if err != nil {
			tx.Rollback()
			return n, err
		}
		if err = tx.Commit(); err != nil {
			return n, err
		}
		slog.Warn("recovered abandoned request", "attempt", d.attempt, "state", "unknown")
	}
	return n, nil
}

func (s *Store) Activity(runID string) ([]Activity, error) {
	if _, err := s.RecoverStale(); err != nil {
		return nil, err
	}
	attempts, err := s.Attempts("", runID)
	if err != nil {
		return nil, err
	}
	out := []Activity{}
	for _, a := range attempts {
		r, err := s.Request(a.Hash)
		if err != nil {
			return nil, err
		}
		if r.AttemptID != a.ID || (r.State != "pending" && r.State != "unknown") {
			continue
		}
		v := Activity{Attempt: a, State: r.State, LocalState: "outcome_unknown"}
		var body openrouter.Request
		if err := json.Unmarshal(r.Body, &body); err != nil {
			return nil, err
		}
		v.Model = body.Model
		v.Effort = body.Reasoning.Effort
		if r.State == "pending" {
			v.LocalState = "unverified"
			locked, err := s.workerLocked(a.RunID)
			if err != nil {
				return nil, err
			}
			var worker string
			err = s.db.QueryRow("SELECT state FROM workers WHERE run_id=?", a.RunID).Scan(&worker)
			if err == nil && worker == "running" && locked {
				v.LocalState = "running"
			}
		}
		if started, err := time.Parse(time.RFC3339Nano, a.Started); err == nil {
			v.SinceDispatchSeconds = time.Since(started).Seconds()
		}
		out = append(out, v)
	}
	return out, nil
}

func (s *Store) StartAttempt(r Request, j Job) error {
	b, err := json.Marshal(AttemptDetails{Question: j.Question, Tier: j.Tier})
	if err != nil {
		return err
	}
	_, err = s.db.Exec("UPDATE attempts SET details=? WHERE id=?", b, r.AttemptID)
	return err
}

func (s *Store) Progress(r Request, p openrouter.Progress) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var raw []byte
	if err = tx.QueryRow("SELECT details FROM attempts WHERE id=?", r.AttemptID).Scan(&raw); err != nil {
		return err
	}
	var d AttemptDetails
	if err = json.Unmarshal(raw, &d); err != nil {
		return err
	}
	if p.GenerationID != "" {
		if d.GenerationID != "" && d.GenerationID != p.GenerationID {
			return fmt.Errorf("generation ID changed for attempt %d", r.AttemptID)
		}
		d.GenerationID = p.GenerationID
	}
	if p.RequestID != "" {
		d.RequestID = p.RequestID
	}
	if p.Provider != "" {
		d.Provider = p.Provider
	}
	d.LastEvent = time.Now().UTC().Format(time.RFC3339Nano)
	raw, err = json.Marshal(d)
	if err != nil {
		return err
	}
	if _, err = tx.Exec("UPDATE attempts SET details=?,http_status=? WHERE id=? AND finished=''", raw, p.HTTPStatus, r.AttemptID); err != nil {
		return err
	}
	if p.Raw != "" {
		if _, err = tx.Exec("INSERT INTO stream_events(attempt_id,received,data) VALUES(?,?,?)", r.AttemptID, d.LastEvent, p.Raw); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ReconcileAttempt attaches accounting to the exact billed attempt, including
// superseded retries. Metadata can never turn an unanswered request into a score.
func (s *Store) ReconcileAttempt(a Attempt, usage openrouter.Usage, raw []byte) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	a.Details.Usage = &usage
	a.Details.Reconciled = true
	b, err := json.Marshal(a.Details)
	if err != nil {
		return err
	}
	if _, err = tx.Exec("UPDATE attempts SET details=? WHERE id=?", b, a.ID); err != nil {
		return err
	}
	if _, err = tx.Exec("INSERT INTO reconciliations(hash,created,response,attempt_id) VALUES(?,?,?,?)", a.Hash, time.Now().UTC().Format(time.RFC3339Nano), raw, a.ID); err != nil {
		return err
	}
	r, _, err := scanRequest(tx.QueryRow("SELECT "+requestColumns+" FROM requests WHERE hash=?", a.Hash))
	if err != nil {
		return err
	}
	if r.AttemptID == a.ID {
		r.Result.Generation.ID = a.Details.GenerationID
		r.Result.Generation.Usage = usage
		if a.Details.Provider != "" {
			r.Result.Generation.Provider = a.Details.Provider
		}
		b, err = json.Marshal(r.Result)
		if err != nil {
			return err
		}
		if _, err = tx.Exec("UPDATE requests SET result=? WHERE hash=? AND attempt_id=?", b, r.Hash, a.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Events(attemptID int64) ([]json.RawMessage, error) {
	rows, err := s.db.Query("SELECT data FROM stream_events WHERE attempt_id=? ORDER BY id", attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []json.RawMessage
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		b, err := json.Marshal(raw)
		if err != nil {
			return nil, err
		}
		events = append(events, b)
	}
	return events, rows.Err()
}
