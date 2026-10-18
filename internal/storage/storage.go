// Package storage keeps request claims, immutable attempts, and reproducible run snapshots.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	_ "modernc.org/sqlite"
)

type Store struct {
	db   *sql.DB
	path string
}
type Selection struct{ Models, Efforts, Questions []string }
type Run struct {
	ID        string             `json:"id"`
	Created   time.Time          `json:"created"`
	Protocol  string             `json:"protocol"`
	Snapshot  dataset.Snapshot   `json:"snapshot"`
	Catalog   openrouter.Catalog `json:"catalog"`
	Selection Selection          `json:"selection"`
	BaseURL   string             `json:"base_url"`
	Timeout   time.Duration      `json:"request_timeout_ns,omitempty"`
	BudgetUSD float64            `json:"budget_usd,omitempty"`
}
type Job struct {
	AttemptID int64  `json:"attempt_id,omitempty"`
	RunID     string `json:"run_id"`
	Question  string `json:"question"`
	Model     string `json:"model"`
	Tier      string `json:"tier"`
	Effort    string `json:"sent_effort,omitempty"`
	Hash      string `json:"request_hash,omitempty"`
	Status    string `json:"status"`
	Note      string `json:"note,omitempty"`
	Cached    bool   `json:"cached"`
}
type Result struct {
	Generation openrouter.Generation `json:"generation"`
	Answer     int                   `json:"answer"`
	Invalid    string                `json:"invalid,omitempty"`
}
type Request struct {
	Hash      string          `json:"hash"`
	Body      json.RawMessage `json:"body"`
	State     string          `json:"state"`
	Result    Result          `json:"result"`
	AttemptID int64           `json:"attempt_id"`
	Note      string          `json:"note,omitempty"`
}
type Attempt struct {
	ID         int64          `json:"id"`
	Hash       string         `json:"request_hash"`
	RunID      string         `json:"run_id"`
	Started    string         `json:"started"`
	Finished   string         `json:"finished,omitempty"`
	HTTPStatus int            `json:"http_status"`
	Response   string         `json:"response"`
	Error      string         `json:"error,omitempty"`
	Details    AttemptDetails `json:"details"`
}

type AttemptDetails struct {
	GenerationID string            `json:"generation_id,omitempty"`
	RequestID    string            `json:"request_id,omitempty"`
	Provider     string            `json:"provider,omitempty"`
	Question     string            `json:"question,omitempty"`
	Tier         string            `json:"tier,omitempty"`
	LastEvent    string            `json:"last_event,omitempty"`
	Reconciled   bool              `json:"reconciled,omitempty"`
	Usage        *openrouter.Usage `json:"usage,omitempty"`
}

func (a Attempt) Usage() openrouter.Usage {
	if a.Details.Usage != nil {
		return *a.Details.Usage
	}
	var g openrouter.Generation
	if a.Finished != "" && a.Error == "" {
		_ = json.Unmarshal([]byte(a.Response), &g)
	}
	return g.Usage
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, path: path}
	_, err = db.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000;
CREATE TABLE IF NOT EXISTS runs (id TEXT PRIMARY KEY, protocol TEXT NOT NULL, data BLOB NOT NULL);
CREATE TABLE IF NOT EXISTS requests (hash TEXT PRIMARY KEY, body BLOB NOT NULL, state TEXT NOT NULL, result BLOB, attempt_id INTEGER NOT NULL DEFAULT 0, note TEXT NOT NULL DEFAULT '', owner INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS attempts (id INTEGER PRIMARY KEY AUTOINCREMENT, hash TEXT NOT NULL REFERENCES requests(hash), run_id TEXT NOT NULL REFERENCES runs(id), started TEXT NOT NULL, finished TEXT NOT NULL DEFAULT '', http_status INTEGER NOT NULL DEFAULT 0, response BLOB, error TEXT NOT NULL DEFAULT '');
CREATE TABLE IF NOT EXISTS jobs (id INTEGER PRIMARY KEY AUTOINCREMENT, run_id TEXT NOT NULL REFERENCES runs(id), data BLOB NOT NULL);
CREATE TABLE IF NOT EXISTS reconciliations (id INTEGER PRIMARY KEY AUTOINCREMENT, hash TEXT NOT NULL REFERENCES requests(hash), created TEXT NOT NULL, response BLOB NOT NULL);`)
	if err != nil {
		db.Close()
		return nil, err
	}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }
func (s *Store) AddRun(r Run) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO runs(id,protocol,data) VALUES(?,?,?)", r.ID, r.Protocol, b)
	return err
}
func (s *Store) Run(id string) (Run, error) {
	var r Run
	var b []byte
	err := s.db.QueryRow("SELECT data FROM runs WHERE id=?", id).Scan(&b)
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(b, &r)
	return r, err
}
func (s *Store) LatestRun() (Run, error) {
	var id string
	if err := s.db.QueryRow("SELECT id FROM runs ORDER BY rowid DESC LIMIT 1").Scan(&id); err != nil {
		return Run{}, err
	}
	return s.Run(id)
}
func (s *Store) AddJob(j Job) error {
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO jobs(run_id,data) VALUES(?,?)", j.RunID, b)
	return err
}
func (s *Store) Jobs(runID, protocol string) ([]Job, error) {
	query := "SELECT j.data FROM jobs j JOIN runs r ON j.run_id=r.id WHERE r.protocol=?"
	args := []any{protocol}
	if runID != "" {
		query += " AND j.run_id=?"
		args = append(args, runID)
	}
	query += " ORDER BY j.id DESC"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []Job
	for rows.Next() {
		var b []byte
		var j Job
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &j); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func scanRequest(row interface{ Scan(...any) error }) (Request, int, error) {
	var r Request
	var data []byte
	var body []byte
	var owner int
	err := row.Scan(&r.Hash, &body, &r.State, &data, &r.AttemptID, &r.Note, &owner)
	if err != nil {
		return r, owner, err
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, &r.Result)
	}
	r.Body = body
	return r, owner, err
}

const requestColumns = "hash,body,state,result,attempt_id,note,owner"

func (s *Store) Request(hash string) (Request, error) {
	r, _, err := scanRequest(s.db.QueryRow("SELECT "+requestColumns+" FROM requests WHERE hash=?", hash))
	return r, err
}

// Claim is atomic across processes. Unknown dispatched requests require an explicit retry.
func (s *Store) Claim(runID, hash string, body []byte, retryUnknown bool) (Request, bool, error) {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return Request{}, false, err
	}
	defer tx.Rollback()
	// A write first obtains SQLite's writer lock before inspecting the claim.
	_, err = tx.Exec("INSERT OR IGNORE INTO requests(hash,body,state,owner) VALUES(?,?,'ready',?)", hash, body, os.Getpid())
	if err != nil {
		return Request{}, false, err
	}
	r, owner, err := scanRequest(tx.QueryRow("SELECT "+requestColumns+" FROM requests WHERE hash=?", hash))
	if err != nil {
		return r, false, err
	}
	if r.State == "pending" && !processAlive(owner) {
		r.State = "unknown"
		r.Note = "previous process exited after claiming the request"
		if _, err := tx.Exec("UPDATE requests SET state='unknown',note=? WHERE hash=?", r.Note, hash); err != nil {
			return r, false, err
		}
		if _, err := tx.Exec("UPDATE attempts SET finished=?,error=? WHERE id=? AND finished=''", time.Now().UTC().Format(time.RFC3339Nano), r.Note, r.AttemptID); err != nil {
			return r, false, err
		}
	}
	allowed := r.State == "ready" || r.State == "retryable" || (r.State == "unknown" && retryUnknown)
	if !allowed {
		return r, false, tx.Commit()
	}
	res, err := tx.Exec("INSERT INTO attempts(hash,run_id,started) VALUES(?,?,?)", hash, runID, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return r, false, err
	}
	r.AttemptID, err = res.LastInsertId()
	if err != nil {
		return r, false, err
	}
	_, err = tx.Exec("UPDATE requests SET state='pending',owner=?,attempt_id=?,note='',result=NULL WHERE hash=?", os.Getpid(), r.AttemptID, hash)
	if err != nil {
		return r, false, err
	}
	r.State = "pending"
	r.Result = Result{}
	return r, true, tx.Commit()
}
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = p.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, os.ErrPermission)
}

func (s *Store) Finish(r Request, state string, response openrouter.HTTPResult, result Result, note string) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE attempts SET finished=?,http_status=?,response=?,error=? WHERE id=?", time.Now().UTC().Format(time.RFC3339Nano), response.Status, []byte(response.Body), response.Error, r.AttemptID)
	if err != nil {
		return err
	}
	res, err := tx.Exec("UPDATE requests SET state=?,result=?,note=? WHERE hash=? AND attempt_id=?", state, b, note, r.Hash, r.AttemptID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("request claim changed for %s", r.Hash)
	}
	return tx.Commit()
}
func (s *Store) Attempts(hash, runID string) ([]Attempt, error) {
	query := "SELECT id,hash,run_id,started,finished,http_status,coalesce(response,''),error,details FROM attempts WHERE 1=1"
	var args []any
	if hash != "" {
		query += " AND hash=?"
		args = append(args, hash)
	}
	if runID != "" {
		query += " AND run_id=?"
		args = append(args, runID)
	}
	query += " ORDER BY id"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Attempt
	for rows.Next() {
		var a Attempt
		var details []byte
		if err := rows.Scan(&a.ID, &a.Hash, &a.RunID, &a.Started, &a.Finished, &a.HTTPStatus, &a.Response, &a.Error, &details); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(details, &a.Details); err != nil {
			return nil, err
		}
		if a.Details.GenerationID == "" {
			var g openrouter.Generation
			if json.Unmarshal([]byte(a.Response), &g) == nil {
				a.Details.GenerationID = g.ID
			}
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
