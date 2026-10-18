package report

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"konkour-bench/internal/bench"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

func ptr[T any](v T) *T { return &v }
func outcome(correct bool) Outcome {
	u := openrouter.Usage{Prompt: ptr(int64(10)), Completion: ptr(int64(30)), Total: ptr(int64(40)), Cost: ptr(0.01)}
	u.CompletionDetails.Reasoning = ptr(int64(20))
	return Outcome{Job: storage.Job{Model: "test/model", Tier: "max", Status: "complete"}, Correct: correct, Usage: u}
}
func TestScoringAndAccounting(t *testing.T) {
	r := Aggregate([]Outcome{outcome(true), outcome(false)})
	if r.Accuracy == nil || *r.Accuracy != 0.5 || r.Correct != 1 || r.Incorrect != 1 || *r.Tokens != 80 || *r.Reasoning != 40 || *r.Cost != 0.02 {
		t.Fatalf("row %+v", r)
	}
	failure := outcome(false)
	failure.Status = "infra_error"
	r = Aggregate([]Outcome{outcome(true), failure})
	if r.Accuracy != nil || r.Status != "incomplete" || r.Total != 2 || r.Failed != 1 {
		t.Fatalf("failure inflated score: %+v", r)
	}
	failure.Status = "skipped"
	r = Aggregate([]Outcome{outcome(true), failure})
	if r.Status != "ineligible" || r.Accuracy != nil {
		t.Fatalf("skip inflated score: %+v", r)
	}
	u := outcome(true)
	u.Usage.Cost = nil
	u.Usage.Total = nil
	r = Aggregate([]Outcome{u})
	if r.Accuracy == nil || r.Cost != nil || r.Tokens != nil {
		t.Fatalf("unknown accounting %+v", r)
	}
}
func TestParetoDominanceTiesAndCohorts(t *testing.T) {
	row := func(a, c float64, tokens int64) Row {
		return Row{Status: "complete", Cohort: "text-only", Accuracy: ptr(a), Cost: ptr(c), Tokens: ptr(tokens)}
	}
	rows := []Row{row(0.9, 1, 100), row(0.8, 2, 200), row(0.9, 1, 100), row(0.7, 0.5, 50), row(1, 0, 0), row(1, 0, 0)}
	rows[4].Cohort = "image-only"
	rows[5].Status = "incomplete"
	Frontiers(rows)
	for _, i := range []int{0, 2, 3, 4} {
		if !rows[i].Pareto || !rows[i].CostFrontier || !rows[i].TokenFrontier {
			t.Fatalf("row %d incorrectly dominated", i)
		}
	}
	if rows[1].Pareto || rows[1].CostFrontier || rows[1].TokenFrontier || rows[5].Pareto {
		t.Fatal("dominated/incomplete row on frontier")
	}
}
func TestArtifactsAndEscaping(t *testing.T) {
	d := Document{Models: []string{`</script><script>alert(1)</script>`}, Categories: []string{"ریاضی"}, Rows: []Row{{Model: "=danger", Tier: "max", Cohort: "text-only", Status: "complete", Accuracy: ptr(0.5), CostPerQuestion: ptr(0.01), TokensPerQuestion: ptr(20.0)}}}
	dir := t.TempDir()
	if err := Write(dir, d); err != nil {
		t.Fatal(err)
	}
	html, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(html), d.Models[0]) {
		t.Fatal("unescaped script content")
	}
	marker := `<script id="data" type="application/json">`
	start := strings.Index(string(html), marker) + len(marker)
	end := strings.Index(string(html)[start:], "</script>")
	if end < 0 || !json.Valid(html[start:start+end]) {
		t.Fatal("embedded report data is not valid JSON")
	}
	b, err := os.ReadFile(filepath.Join(dir, "results.json"))
	if err != nil || !json.Valid(b) {
		t.Fatalf("invalid JSON: %v", err)
	}
	csv, err := os.ReadFile(filepath.Join(dir, "leaderboard.csv"))
	if err != nil || !strings.Contains(string(csv), "'=danger") {
		t.Fatalf("unsafe CSV: %v", err)
	}
	for _, name := range []string{"accuracy.svg", "cost.svg", "tokens.svg"} {
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		decoder := xml.NewDecoder(f)
		for {
			_, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("invalid %s: %v", name, err)
			}
		}
		f.Close()
	}
}

func TestReportRescoresWithoutInferenceAndPreservesHistory(t *testing.T) {
	snap, err := dataset.Load("../../benchmark.sample.json")
	if err != nil {
		t.Fatal(err)
	}
	var questions []dataset.Question
	for _, q := range snap.Questions {
		if q.Number == 2 || q.Number == 9 {
			questions = append(questions, q)
		}
	}
	snap.Questions = questions
	snap.Config.Models = []string{"openai/gpt-6-luna"}
	sel := storage.Selection{Models: snap.Config.Models, Efforts: []string{"medium"}}
	s, err := storage.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	base := "https://test.invalid/api/v1"
	if err := s.AddRun(storage.Run{ID: "original", Protocol: bench.Protocol(), Snapshot: snap, Selection: sel, BaseURL: base}); err != nil {
		t.Fatal(err)
	}
	for _, q := range snap.Questions {
		body, err := openrouter.Build(q, sel.Models[0], "medium")
		if err != nil {
			t.Fatal(err)
		}
		hash := bench.RequestHash(base, body)
		req, claimed, err := s.Claim("original", hash, body, false)
		if err != nil || !claimed {
			t.Fatalf("claim: %v %v", claimed, err)
		}
		usage := outcome(true).Usage
		raw, _ := json.Marshal(map[string]any{"usage": usage})
		if err := s.Finish(req, "complete", openrouter.HTTPResult{Status: 200, Body: raw}, storage.Result{Generation: openrouter.Generation{Usage: usage}, Answer: q.Correct}, ""); err != nil {
			t.Fatal(err)
		}
		if err := s.AddJob(storage.Job{RunID: "original", Question: q.ID(), Model: sel.Models[0], Tier: "medium", Effort: "medium", Hash: hash, Status: "complete"}); err != nil {
			t.Fatal(err)
		}
	}
	check := func(doc Document, want *float64) {
		t.Helper()
		for _, r := range doc.Rows {
			if r.Cohort == "full" && r.Dataset == "" && r.Category == "" {
				if want == nil {
					if r.Accuracy != nil || r.Status != "incomplete" {
						t.Fatalf("wanted incomplete row: %+v", r)
					}
				} else if r.Accuracy == nil || *r.Accuracy != *want || r.Total != 2 {
					t.Fatalf("wrong score: %+v", r)
				}
				return
			}
		}
		t.Fatal("missing overall full cohort")
	}
	doc, err := Build(s, snap, "original", base, sel)
	if err != nil {
		t.Fatal(err)
	}
	check(doc, ptr(1.0))
	if len(doc.Spending) != 1 || doc.Spending[0].Attempts != 2 || doc.Spending[0].KnownCost != 0.02 {
		t.Fatalf("spending %+v", doc.Spending)
	}
	snap.Questions[0].Correct = 1
	rescored, err := Build(s, snap, "", base, sel)
	if err != nil {
		t.Fatal(err)
	}
	check(rescored, ptr(0.5))
	if rescored.ScoringRevision == doc.ScoringRevision {
		t.Fatal("key change did not update scoring revision")
	}
	original, err := s.Run("original")
	if err != nil {
		t.Fatal(err)
	}
	historical, err := Build(s, original.Snapshot, "original", base, sel)
	if err != nil {
		t.Fatal(err)
	}
	check(historical, ptr(1.0))
	snap.Questions[1].Asset.Bytes = append(snap.Questions[1].Asset.Bytes, 0)
	changed, err := Build(s, snap, "", base, sel)
	if err != nil {
		t.Fatal(err)
	}
	check(changed, nil)
	attempts, err := s.Attempts("", "")
	if err != nil || len(attempts) != 2 {
		t.Fatalf("rescore created inference attempts: %d %v", len(attempts), err)
	}
}

func TestStreamHistoryAndInterruptedCostStaySeparate(t *testing.T) {
	snap := dataset.Snapshot{Config: dataset.Config{Models: []string{"test/model"}}, Questions: []dataset.Question{{Dataset: "exam", Number: 1, Text: "question", Correct: 4}}}
	sel := storage.Selection{Models: snap.Config.Models, Efforts: []string{"max"}}
	s, err := storage.Open(filepath.Join(t.TempDir(), "bench.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	base := "https://test.invalid"
	if err := s.AddRun(storage.Run{ID: "run", Protocol: bench.Protocol(), Snapshot: snap, Selection: sel, BaseURL: base}); err != nil {
		t.Fatal(err)
	}
	body, err := openrouter.BuildStream(snap.Questions[0], "test/model", "max")
	if err != nil {
		t.Fatal(err)
	}
	hash := bench.RequestHash(base, body)
	req, _, err := s.Claim("run", hash, body, false)
	if err != nil {
		t.Fatal(err)
	}
	j := storage.Job{RunID: "run", Question: "exam:1", Model: "test/model", Tier: "max", Effort: "max", Hash: hash, Status: "pending", AttemptID: req.AttemptID}
	if err := s.Finish(req, "unknown", openrouter.HTTPResult{Error: "disconnected"}, storage.Result{}, "unknown"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddJob(j); err != nil {
		t.Fatal(err)
	}
	attempts, err := s.Attempts(hash, "")
	if err != nil {
		t.Fatal(err)
	}
	a := attempts[0]
	a.Details.GenerationID = "gen-interrupted"
	if err := s.ReconcileAttempt(a, openrouter.Usage{Cost: ptr(0.11)}, []byte(`{"data":{"cancelled":true}}`)); err != nil {
		t.Fatal(err)
	}
	doc, err := Build(s, snap, "run", base, sel)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Outcomes) != 1 || doc.Outcomes[0].Status != "infra_error" || doc.Outcomes[0].Answer != 0 || doc.Spending[0].KnownCost != 0.11 || doc.Spending[0].UnknownCost != 0 {
		t.Fatalf("interrupted accounting/scoring %+v %+v", doc.Outcomes, doc.Spending)
	}
	for _, row := range doc.Rows {
		if row.Accuracy != nil || row.Status != "incomplete" {
			t.Fatal("interrupted answer ranked", row)
		}
	}
	retry, _, err := s.Claim("run", hash, body, true)
	if err != nil {
		t.Fatal(err)
	}
	result := storage.Result{Answer: 4, Generation: openrouter.Generation{Usage: openrouter.Usage{Cost: ptr(0.2)}}}
	raw, err := json.Marshal(result.Generation)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Finish(retry, "complete", openrouter.HTTPResult{Status: 200, Body: raw}, result, ""); err != nil {
		t.Fatal(err)
	}
	// A later retry cannot retroactively complete the original pending job.
	doc, err = Build(s, snap, "run", base, sel)
	if err != nil || doc.Outcomes[0].Status != "infra_error" {
		t.Fatalf("old attempt upgraded by retry: %+v %v", doc.Outcomes, err)
	}
	// Conversely, a crash after persisting the final response but before the
	// final job event must not discard that same attempt's completed answer.
	j.AttemptID = retry.AttemptID
	if err := s.AddJob(j); err != nil {
		t.Fatal(err)
	}
	doc, err = Build(s, snap, "run", base, sel)
	if err != nil {
		t.Fatal(err)
	}
	if !doc.Outcomes[0].Correct || doc.Spending[0].KnownRetryCost != 0.11 || doc.Spending[0].KnownCost != 0.31 {
		t.Fatalf("retry accounting %+v %+v", doc.Outcomes, doc.Spending)
	}
}
