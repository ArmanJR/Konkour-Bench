package report

import (
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"konkour-bench/internal/bench"
	"konkour-bench/internal/dataset"
	"konkour-bench/internal/openrouter"
	"konkour-bench/internal/storage"
)

func TestQuestionIntervalsAndResourceQuantiles(t *testing.T) {
	if Wilson(0, 0) != nil {
		t.Fatal("empty sample has an interval")
	}
	ci := Wilson(50, 100)
	if math.Abs(ci.Low-.4038315304) > 1e-9 || math.Abs(ci.High-.5961684696) > 1e-9 {
		t.Fatalf("Wilson interval: %+v", ci)
	}
	if Wilson(0, 80).Low > 1e-15 || Wilson(80, 80).High < 1-1e-15 {
		t.Fatal("boundary intervals")
	}
	values := []float64{30, 0, 20, 10}
	d := Describe(values)
	if d.N != 4 || *d.Median != 15 || *d.P90 != 27 || *d.Max != 30 || values[0] != 30 {
		t.Fatalf("quantiles: %+v; input: %v", d, values)
	}
	if Describe(nil).Median != nil || *Describe([]float64{0}).Median != 0 {
		t.Fatal("unknown and zero confused")
	}
}

func TestIncompleteBoundsAndKnownResourceSamples(t *testing.T) {
	good, missing := outcome(true), outcome(false)
	good.Duration = ptr(0.0)
	missing.Status = "infra_error"
	r := Aggregate([]Outcome{good, missing})
	if r.Accuracy != nil || r.Interval != nil || r.Bounds == nil || r.Bounds.Low != .5 || r.Bounds.High != 1 {
		t.Fatalf("incomplete ranked: %+v", r)
	}
	if r.Cost != nil || r.CostDistribution.N != 1 || *r.CostDistribution.Median != .01 || r.Latency.N != 1 || *r.Latency.Median != 0 {
		t.Fatalf("partial accounting: %+v", r)
	}
	missing.Status = "skipped"
	if Aggregate([]Outcome{good, missing}).Bounds != nil {
		t.Fatal("capability exclusion has completion bounds")
	}
	if Aggregate(nil).Accuracy != nil {
		t.Fatal("empty group scored")
	}
}

func analysisFixture() Document {
	d := Document{}
	for _, config := range []struct {
		model, tier string
		answers     []int
	}{
		{"test/a", "medium", []int{1, 2, 1}},
		{"test/a", "max", []int{2, 1, 0}},
		{"test/b", "medium", []int{2, 2, 1}},
		{"test/c", "medium", []int{2, 2, 1}},
	} {
		var group []Outcome
		for i, answer := range config.answers {
			o := outcome(answer == 1)
			o.Model = config.model
			o.Tier = config.tier
			o.Question = "exam:" + string(rune('1'+i))
			o.Dataset = "exam"
			o.Category = "subject"
			o.Expected = 1
			o.Answer = answer
			o.Image = i == 2
			if answer == 0 {
				o.Status = "infra_error"
				o.Note = "timeout"
			}
			if config.tier == "max" {
				o.Usage.Cost = ptr(.02)
				o.Usage.Total = ptr(int64(80))
			}
			group = append(group, o)
			d.Outcomes = append(d.Outcomes, o)
		}
		r := Aggregate(group)
		r.Cohort = "full"
		d.Rows = append(d.Rows, r)
		r = Aggregate(group[:2])
		r.Cohort = "text-only"
		d.Rows = append(d.Rows, r)
	}
	d.Views = Analyze(d)
	return d
}

func TestAnalysisFixedPanelAndPairedEffort(t *testing.T) {
	d := analysisFixture()
	var full, text AnalysisView
	for _, v := range d.Views {
		if v.Cohort == "full" {
			full = v
		} else {
			text = v
		}
	}
	if full.Coverage.Questions != 3 || full.Coverage.Images != 1 || full.Coverage.Complete != 11 || full.Coverage.Ranked != 3 || text.Coverage.Ranked != 4 {
		t.Fatalf("coverage: %+v / %+v", full.Coverage, text.Coverage)
	}
	if len(full.Efforts) != 1 {
		t.Fatalf("pairs: %+v", full.Efforts)
	}
	e := full.Efforts[0]
	if e.Paired != 2 || e.Total != 3 || e.Gained != 1 || e.Lost != 1 || *e.Delta != 0 || *e.CostRatio != 2 || *e.TokenRatio != 2 {
		t.Fatalf("not paired on same questions: %+v", e)
	}
	for _, q := range full.Questions {
		if q.Panel != 3 {
			t.Fatal("incomplete config leaked into panel", q)
		}
		if q.ID == "exam:1" && (!q.Review || q.Correct != 1 || q.Answers[2] != 2) {
			t.Fatal("review flag", q)
		}
	}
	// Unknown subscription charges must not become a zero multiplier.
	for i := range d.Outcomes {
		d.Outcomes[i].Usage.Cost = nil
	}
	for _, v := range Analyze(d) {
		for _, e := range v.Efforts {
			if e.CostRatio != nil {
				t.Fatal("unknown costs used")
			}
		}
	}
}

func TestReviewRequiresMajorityAndInvalidIsNotAnOption(t *testing.T) {
	d := analysisFixture()
	for i := range d.Outcomes {
		if d.Outcomes[i].Question == "exam:1" {
			d.Outcomes[i].Invalid = "malformed"
			d.Outcomes[i].Correct = false
			d.Outcomes[i].Answer = 0
		}
	}
	for _, v := range Analyze(d) {
		for _, q := range v.Questions {
			if q.ID == "exam:1" && (q.Review || q.Answers[0] != q.Panel) {
				t.Fatal("invalid treated as consensus answer", q)
			}
		}
	}
}

func TestAnalysisExportsPreserveNullsAndEscapeLabels(t *testing.T) {
	d := analysisFixture()
	d.Outcomes[0].Model = " \t=HYPERLINK(\"bad\")"
	d.Outcomes[0].Usage.Cost = nil
	dir := t.TempDir()
	if err := Write(dir, d); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"leaderboard.csv", "outcomes.csv", "questions.csv", "effort.csv", "ranks.csv", "separation.csv", "length.csv", "estimates.csv"} {
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		records, err := csv.NewReader(f).ReadAll()
		f.Close()
		if err != nil {
			t.Fatal(name, err)
		}
		if len(records) < 2 {
			t.Fatal("empty export", name)
		}
		if name == "leaderboard.csv" && (records[0][18] != "accuracy_ci_low" || records[1][18] == "" || records[0][31] != "latency_seconds_p90") {
			t.Fatal("missing analysis columns", records[0], records[1])
		}
		if name == "outcomes.csv" && (records[1][3] != "'"+d.Outcomes[0].Model || records[1][17] != "") {
			t.Fatal("unsafe label or unknown converted to zero", records[1])
		}
		if name == "estimates.csv" && (records[0][16] != "check_predicted" || records[1][4] != "max" || records[1][8] != "1") {
			t.Fatal("completion estimate export", records[0], records[1])
		}
	}
	html, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"name=\"description\"", "id=\"effort\"", "id=\"questions\"", "data:font/woff2;base64,", "renderMathInElement", "const reportLogos =", "data:image/svg+xml;base64,", "Copyright (c) 2023 LobeHub", "id=\"chart-fonts\"", "font-family:Inter", "SIL OPEN FONT LICENSE"} {
		if !strings.Contains(string(html), required) {
			t.Fatal("missing offline feature", required)
		}
	}
	if strings.Contains(string(html), "<script src=") || strings.Contains(string(html), "<link rel=\"stylesheet\"") {
		t.Fatal("external runtime asset")
	}
}

func TestWriteErrorsPreserveExistingReport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.html")
	if err := os.WriteFile(path, []byte("previous report"), 0600); err != nil {
		t.Fatal(err)
	}
	d := Document{Rows: []Row{{Cost: ptr(math.NaN())}}}
	if err := Write(dir, d); err == nil || !strings.Contains(err.Error(), "encode report") {
		t.Fatalf("invalid data error: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != "previous report" {
		t.Fatal("failed rendering overwrote report", err)
	}
	if err := Write(path, Document{}); err == nil {
		t.Fatal("writing into a file unexpectedly succeeded")
	}
}

func TestCachedAnswersUseOriginalTimingAndOnlySelectedQuestionContent(t *testing.T) {
	s, err := storage.Open(filepath.Join(t.TempDir(), "bench.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	q := dataset.Question{Dataset: "exam", Number: 1, Text: "question", Correct: 1}
	snap := dataset.Snapshot{Questions: []dataset.Question{q, {Dataset: "exam", Number: 2, Text: "not selected", Correct: 2}}}
	sel := storage.Selection{Models: []string{"test/model"}, Efforts: []string{"medium"}, Questions: []string{q.ID()}}
	base := "https://test.invalid"
	for _, id := range []string{"origin", "resume"} {
		if err := s.AddRun(storage.Run{ID: id, Created: time.Now().UTC(), Protocol: bench.Protocol(), Snapshot: snap, Selection: sel, BaseURL: base}); err != nil {
			t.Fatal(err)
		}
	}
	body, err := openrouter.Build(q, "test/model", "medium")
	if err != nil {
		t.Fatal(err)
	}
	hash := bench.RequestHash(base, body)
	req, _, err := s.Claim("origin", hash, body, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Finish(req, "complete", openrouter.HTTPResult{Status: 200, Body: []byte(`{}`)}, storage.Result{Answer: 1}, ""); err != nil {
		t.Fatal(err)
	}
	if err := s.AddJob(storage.Job{RunID: "resume", Question: q.ID(), Model: "test/model", Tier: "medium", Effort: "medium", Hash: hash, Status: "complete", Cached: true}); err != nil {
		t.Fatal(err)
	}
	d, err := Build(s, snap, "resume", base, sel)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.QuestionBank) != 1 || d.QuestionBank[0].ID != q.ID() || d.Outcomes[0].OriginRun != "origin" || d.Outcomes[0].Duration == nil || len(d.Sources) != 2 {
		t.Fatalf("cached provenance/content: %+v %+v", d.Outcomes, d.Sources)
	}
	if len(d.Spending) != 1 || d.Spending[0].Attempts != 0 {
		t.Fatal("cached inference added spending", d.Spending)
	}
}
